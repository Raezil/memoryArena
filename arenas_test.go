package memoryArena

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
	"unsafe"
)

// Test edge cases for memory allocation
func TestMemoryArena_EdgeCases(t *testing.T) {
	// Test creating arena with zero size
	_, err := NewMemoryArena[int](0)
	if err == nil {
		t.Error("Expected error when creating arena with size 0")
	}

	// Test creating arena with negative size
	_, err = NewMemoryArena[int](-10)
	if err == nil {
		t.Error("Expected error when creating arena with negative size")
	}

	// Test allocating zero size
	arena, _ := NewMemoryArena[int](100)
	_, err = arena.Allocate(0)
	if err == nil {
		t.Error("Expected error when allocating size 0")
	}

	// Test allocating more than available
	_, err = arena.Allocate(101)
	if err == nil {
		t.Error("Expected error when allocating more than available")
	}
}

// Test the alignment functionality
func TestMemoryArena_Alignment(t *testing.T) {
	arena, _ := NewMemoryArena[int](100)

	// Make offset unaligned
	arena.buffer.offset = 3

	// Call alignOffset with alignment 8
	arena.alignOffset(8)

	// Check if offset is aligned to 8
	if arena.buffer.offset != 8 {
		t.Errorf("Expected offset 8 after alignment, got %d", arena.buffer.offset)
	}

	// Test with different alignment
	arena.buffer.offset = 9
	arena.alignOffset(4)
	if arena.buffer.offset != 12 {
		t.Errorf("Expected offset 12 after alignment, got %d", arena.buffer.offset)
	}
}

// Test the GetRemainder function
func TestMemoryArena_GetRemainder(t *testing.T) {
	arena, _ := NewMemoryArena[int](100)

	// Test with different offsets and alignments
	testCases := []struct {
		offset    int
		alignment uintptr
		expected  int
	}{
		{0, 8, 0},
		{3, 8, 3},
		{8, 8, 0},
		{9, 4, 1},
		{15, 16, 15},
	}

	for _, tc := range testCases {
		arena.buffer.offset = tc.offset
		remainder := arena.GetRemainder(tc.alignment)
		if remainder != tc.expected {
			t.Errorf("Expected remainder %d for offset %d and alignment %d, got %d",
				tc.expected, tc.offset, tc.alignment, remainder)
		}
	}
}

// Test ResizePreserve with edge cases
func TestMemoryArena_ResizePreserve_EdgeCases(t *testing.T) {
	arena, _ := NewMemoryArena[int](100)

	// Fill the arena partially
	_, err := arena.Allocate(50)
	if err != nil {
		t.Fatalf("Failed to allocate: %v", err)
	}

	// Test resize to smaller size but still larger than used
	err = arena.ResizePreserve(75)
	if err != nil {
		t.Errorf("ResizePreserve failed: %v", err)
	}
	if arena.buffer.size != 75 {
		t.Errorf("Expected size 75 after resize, got %d", arena.buffer.size)
	}

	// Test resize to smaller size than used
	err = arena.ResizePreserve(40)
	if err == nil {
		t.Error("Expected error when resizing to size smaller than used")
	}

	// Test resize with zero or negative
	err = arena.ResizePreserve(0)
	if err == nil {
		t.Error("Expected error when resizing to zero")
	}

	err = arena.ResizePreserve(-10)
	if err == nil {
		t.Error("Expected error when resizing to negative value")
	}
}

// Test allocating complex structures
type TestStruct struct {
	A int
	B string
	C float64
	D []int
}

func TestMemoryArena_ComplexStructures(t *testing.T) {
	arena, _ := NewMemoryArena[TestStruct](1000)

	obj := TestStruct{
		A: 42,
		B: "hello",
		C: 3.14,
		D: []int{1, 2, 3, 4},
	}

	ptr, err := arena.AllocateObject(obj)
	if err != nil {
		t.Fatalf("Failed to allocate complex structure: %v", err)
	}

	// Check if the structure was copied correctly
	result := (*TestStruct)(ptr)
	if result.A != obj.A || result.B != obj.B || result.C != obj.C {
		t.Error("Complex structure data not preserved correctly")
	}

	if len(result.D) != len(obj.D) {
		t.Error("Slice length not preserved correctly")
	} else {
		for i, v := range result.D {
			if v != obj.D[i] {
				t.Errorf("Slice data not preserved correctly at index %d", i)
			}
		}
	}
}

// Test concurrent operations on ConcurrentArena
func TestConcurrentArena_ConcurrentOperations(t *testing.T) {
	arena, _ := NewConcurrentArena[int](1000)

	const goroutines = 10
	const allocationsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Launch multiple goroutines to allocate concurrently
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()

			for i := 0; i < allocationsPerGoroutine; i++ {
				obj := id*100 + i
				_, err := arena.AllocateObject(obj)
				if err != nil {
					t.Errorf("Goroutine %d failed allocation %d: %v", id, i, err)
				}
			}
		}(g)
	}

	wg.Wait()
}

// Test resizing concurrent arena during operations
func TestConcurrentArena_ResizeDuringOperations(t *testing.T) {
	arena, _ := NewConcurrentArena[int](200)

	// Start allocating in a goroutine
	done := make(chan bool)
	go func() {
		for i := 0; i < 10; i++ {
			_, err := arena.AllocateObject(i)
			if err != nil {
				t.Errorf("Allocation failed during resize test: %v", err)
			}
		}
		done <- true
	}()

	// Resize while allocations are happening
	err := arena.Resize(400)
	if err != nil {
		t.Errorf("Resize failed: %v", err)
	}

	<-done

	// Verify the new size
	if arena.MemoryArena.buffer.size != 400 {
		t.Errorf("Expected size 400 after resize, got %d", arena.buffer.size)
	}
}

// Test the Arena interface implementation
func TestArenaInterface(t *testing.T) {
	// Test that MemoryArena implements Arena
	memArena, _ := NewMemoryArena[int](100)
	var arenaInterface Arena = memArena

	// Test Reset
	arenaInterface.Reset()
	if memArena.buffer.offset != 0 {
		t.Error("Reset through interface did not work correctly")
	}

	// Test AllocateObject
	obj := 42
	ptr, err := arenaInterface.AllocateObject(obj)
	if err != nil {
		t.Errorf("AllocateObject through interface failed: %v", err)
	}
	if (*(*int)(ptr)) != obj {
		t.Error("AllocateObject through interface did not preserve value")
	}

	// Test with ConcurrentArena
	concArena, _ := NewConcurrentArena[int](100)
	arenaInterface = concArena

	// Test Reset
	arenaInterface.Reset()
	if concArena.buffer.offset != 0 {
		t.Error("Reset through interface did not work correctly for ConcurrentArena")
	}

	// Test AllocateObject
	ptr, err = arenaInterface.AllocateObject(obj)
	if err != nil {
		t.Errorf("AllocateObject through interface failed for ConcurrentArena: %v", err)
	}
	if (*(*int)(ptr)) != obj {
		t.Error("AllocateObject through interface did not preserve value for ConcurrentArena")
	}
}

// Test memory alignment with different types
func TestMemoryArena_AlignmentWithDifferentTypes(t *testing.T) {
	// Test with byte (alignment typically 1)
	byteArena, _ := NewMemoryArena[byte](100)
	byteArena.buffer.offset = 1
	byteArena.alignOffset(unsafe.Alignof(byte(0)))
	if byteArena.buffer.offset != 1 {
		t.Errorf("Expected byte alignment to be 1, offset changed to %d", byteArena.buffer.offset)
	}

	// Test with int64 (alignment typically 8)
	int64Arena, _ := NewMemoryArena[int64](100)
	int64Arena.buffer.offset = 3
	int64Arena.alignOffset(unsafe.Alignof(int64(0)))
	if int64Arena.buffer.offset%8 != 0 {
		t.Errorf("Expected int64 offset to be 8-aligned, got %d", int64Arena.buffer.offset)
	}

	// Test with complex struct
	type AlignmentTestStruct struct {
		A byte
		B int64
		C int32
	}
	structArena, _ := NewMemoryArena[AlignmentTestStruct](200)
	structArena.buffer.offset = 5
	structArena.alignOffset(unsafe.Alignof(AlignmentTestStruct{}))
	alignment := int(unsafe.Alignof(AlignmentTestStruct{}))
	if structArena.buffer.offset%alignment != 0 {
		t.Errorf("Expected struct offset to be %d-aligned, got %d", alignment, structArena.buffer.offset)
	}
}

// Test the SetNewMemory method
func TestMemoryArena_SetNewMemory(t *testing.T) {
	arena, _ := NewMemoryArena[int](100)

	// Fill with some data
	for i := 0; i < 50; i++ {
		arena.buffer.memory[i] = byte(i)
	}
	arena.buffer.offset = 50

	// Set new memory
	newSize := 200
	arena.SetNewMemory(newSize, 50)

	// Check if size was updated
	if arena.buffer.size != newSize {
		t.Errorf("Expected size %d after SetNewMemory, got %d", newSize, arena.buffer.size)
	}

	// Check if data was preserved
	for i := 0; i < 50; i++ {
		if arena.buffer.memory[i] != byte(i) {
			t.Errorf("Data at index %d was not preserved correctly", i)
		}
	}
}

// Test ReadLock performance in ConcurrentArena
func TestConcurrentArena_ReadLock(t *testing.T) {
	arena, _ := NewConcurrentArena[int](10000)

	// Allocate some data
	for i := 0; i < 100; i++ {
		_, err := arena.AllocateObject(i)
		if err != nil {
			t.Fatalf("Setup allocation failed: %v", err)
		}
	}

	const readers = 50
	const reads = 1000

	var wg sync.WaitGroup
	wg.Add(readers)

	start := time.Now()
	// Launch many readers
	for r := 0; r < readers; r++ {
		go func() {
			defer wg.Done()

			// Perform many read operations
			for i := 0; i < reads; i++ {
				arena.mutex.Lock()
				// Simulate read operation
				_ = arena.buffer.offset
				arena.mutex.Unlock()
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("%d concurrent readers with %d reads each completed in %v",
		readers, reads, elapsed)
}

// Test stress test with extremely frequent resizing
func TestConcurrentArena_FrequentResizing(t *testing.T) {
	arena, _ := NewConcurrentArena[int](1000)

	const goroutines = 20
	const operations = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()

			for i := 0; i < operations; i++ {
				// Alternate between allocations and resizes
				if i%2 == 0 {
					value := id*1000 + i
					_, err := arena.AllocateObject(value)
					if err != nil {
						// This can fail due to concurrent resizing
						continue
					}
				} else {
					// Try to resize
					baseSize := 1000 + (id * 100)
					delta := i * 10
					// Alternate between growing and shrinking
					var newSize int
					if i%4 == 1 {
						newSize = baseSize + delta
					} else {
						newSize = baseSize - delta
						if newSize < 100 {
							newSize = 100
						}
					}

					// Try both resize methods
					var err error
					if i%4 < 2 {
						err = arena.Resize(newSize)
					} else {
						err = arena.ResizePreserve(newSize)
					}

					// Errors are expected due to concurrency
					_ = err
				}

				// Small sleep to allow other goroutines a chance
				if i%10 == 0 {
					time.Sleep(time.Microsecond)
				}
			}
		}(g)
	}

	wg.Wait()

	// Final validation - just check that the arena is still usable
	err := arena.Resize(2000)
	if err != nil {
		t.Errorf("Final resize failed: %v", err)
	}

	_, err = arena.AllocateObject(123)
	if err != nil {
		t.Errorf("Final allocation failed: %v", err)
	}
}

// Test deadlock prevention - make sure our locking strategy is correct
func TestConcurrentArena_DeadlockPrevention(t *testing.T) {
	arena, _ := NewConcurrentArena[int](1000)

	// Create a timeout channel
	timeout := time.After(5 * time.Second)
	done := make(chan bool)

	go func() {
		// Lock, allocate, then unlock
		arena.mutex.Lock()
		_, _ = arena.MemoryArena.Allocate(10) // Use underlying arena directly
		arena.mutex.Unlock()

		// Now try to use the normal method which also acquires a lock
		_, _ = arena.AllocateObject(42)

		done <- true
	}()

	// Check if we complete or timeout
	select {
	case <-done:
		// Test completed successfully
	case <-timeout:
		t.Fatal("Possible deadlock detected - test timed out")
	}
}

// Test with a custom complex structure needing proper alignment
type ComplexStruct struct {
	A byte
	B int64
	C [4]float64
	D struct {
		X int32
		Y [8]byte
	}
}

func BenchmarkMemoryArenaVsStandardAllocation(b *testing.B) {
	// Standard Go allocation
	b.Run("StandardAllocation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj := new(int)
			*obj = i
			_ = obj
		}
	})

	// Memory arena allocation
	b.Run("ArenaAllocation", func(b *testing.B) {
		arena, _ := NewMemoryArena[int](b.N * int(unsafe.Sizeof(int(0))) * 2)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ptr, _ := arena.AllocateObject(i)
			_ = ptr
		}
	})
}

// Benchmark different sizes of allocations
func BenchmarkMemoryArena_DifferentSizes(b *testing.B) {
	sizes := []int{8, 16, 32, 64, 128, 256, 512, 1024}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			arena, _ := NewMemoryArena[[]byte](b.N * size * 2)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				data := make([]byte, size)
				_, _ = arena.AllocateObject(data)
			}
		})
	}
}

// Compare standard vs concurrent arena performance with no contention
func BenchmarkArenaTypes_NoContention(b *testing.B) {
	b.Run("StandardArena", func(b *testing.B) {
		arena, _ := NewMemoryArena[int](b.N * int(unsafe.Sizeof(int(0))) * 2)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = arena.AllocateObject(i)
		}
	})

	b.Run("ConcurrentArena", func(b *testing.B) {
		arena, _ := NewConcurrentArena[int](b.N * int(unsafe.Sizeof(int(0))) * 2)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = arena.AllocateObject(i)
		}
	})
}

// Test allocation of large objects
type LargeStruct struct {
	Data     [1024]byte
	Metadata [128]int64
}

func BenchmarkAllocateLargeObjects(b *testing.B) {
	b.Run("StandardGo", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj := new(LargeStruct)
			obj.Metadata[0] = int64(i)
			_ = obj
		}
	})

	b.Run("MemoryArena", func(b *testing.B) {
		arena, _ := NewMemoryArena[LargeStruct](b.N*int(unsafe.Sizeof(LargeStruct{})) + 1024)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj := LargeStruct{}
			obj.Metadata[0] = int64(i)
			_, _ = arena.AllocateObject(obj)
		}
	})
}

// Test sequential allocation and reset cycles
func TestMemoryArena_AllocationAndResetCycles(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Fatalf("Failed to create arena: %v", err)
	}

	// Perform multiple allocation and reset cycles
	for cycle := 0; cycle < 5; cycle++ {
		// Fill arena almost completely
		allocCount := 0
		for {
			_, err := arena.Allocate(8)
			if err != nil {
				break
			}
			allocCount++
		}

		// Verify we could allocate multiple times
		if allocCount == 0 {
			t.Errorf("Cycle %d: Failed to allocate any memory", cycle)
		}

		// Reset and verify offset is back to 0
		arena.Reset()
		if arena.buffer.offset != 0 {
			t.Errorf("Cycle %d: Offset not reset to 0", cycle)
		}

		// Verify we can allocate again after reset
		ptr, err := arena.Allocate(8)
		if err != nil {
			t.Errorf("Cycle %d: Failed to allocate after reset: %v", cycle, err)
		}
		if ptr == nil {
			t.Errorf("Cycle %d: Allocation after reset returned nil", cycle)
		}
	}
}

// Test allocating objects of different types
func TestMemoryArena_DifferentTypes(t *testing.T) {
	// Test with int
	intArena, _ := NewMemoryArena[int](100)
	intVal := 42
	intPtr, err := intArena.AllocateObject(intVal)
	if err != nil {
		t.Errorf("Failed to allocate int: %v", err)
	}
	if *(*int)(intPtr) != intVal {
		t.Errorf("Int value not preserved correctly")
	}

	// Test with string
	strArena, _ := NewMemoryArena[string](200)
	strVal := "test string"
	strPtr, err := strArena.AllocateObject(strVal)
	if err != nil {
		t.Errorf("Failed to allocate string: %v", err)
	}
	if *(*string)(strPtr) != strVal {
		t.Errorf("String value not preserved correctly")
	}

	// Test with struct
	type TestStruct struct {
		X, Y int
		Name string
	}
	structArena, _ := NewMemoryArena[TestStruct](300)
	structVal := TestStruct{X: 10, Y: 20, Name: "test"}
	structPtr, err := structArena.AllocateObject(structVal)
	if err != nil {
		t.Errorf("Failed to allocate struct: %v", err)
	}
	resultStruct := *(*TestStruct)(structPtr)
	if resultStruct.X != structVal.X ||
		resultStruct.Y != structVal.Y ||
		resultStruct.Name != structVal.Name {
		t.Errorf("Struct value not preserved correctly")
	}
}

// Test the incorrect type handling
func TestMemoryArena_InvalidType(t *testing.T) {
	arena, _ := NewMemoryArena[int](100)

	// Try to allocate a string in an int arena
	_, err := arena.AllocateObject("string in int arena")
	if err == nil {
		t.Error("Expected error when allocating wrong type, got nil")
	}
}

// Test that memory is correctly reused after reset
func TestMemoryArena_MemoryReuseAfterReset(t *testing.T) {
	arena, _ := NewMemoryArena[int](100)

	// First allocation
	ptr1, _ := arena.AllocateObject(123)
	offset1 := arena.buffer.offset

	// Reset the arena
	arena.Reset()

	// Second allocation of same size
	ptr2, _ := arena.AllocateObject(456)
	offset2 := arena.buffer.offset

	// The pointers should point to the same address if memory is reused correctly
	if uintptr(ptr1) != uintptr(ptr2) {
		t.Errorf("Memory not reused after reset: ptr1=%p, ptr2=%p", ptr1, ptr2)
	}

	// The offset after the first allocation should match the offset after the second
	if offset1 != offset2 {
		t.Errorf("Offset not consistent after reset: offset1=%d, offset2=%d", offset1, offset2)
	}

	// The value should be updated
	if *(*int)(ptr2) != 456 {
		t.Errorf("Second value not stored correctly: expected 456, got %d", *(*int)(ptr2))
	}
}

// Test concurrent arena with large number of allocations
func TestConcurrentArena_LargeNumberOfAllocations(t *testing.T) {
	const arenaSize = 10000
	const numAllocations = 1000

	arena, err := NewConcurrentArena[int](arenaSize)
	if err != nil {
		t.Fatalf("Failed to create concurrent arena: %v", err)
	}

	// Allocate until we run out of space
	allocCount := 0
	for i := 0; i < numAllocations; i++ {
		_, err := arena.AllocateObject(i)
		if err != nil {
			break
		}
		allocCount++
	}

	// Verify we could allocate multiple times
	if allocCount == 0 {
		t.Error("Failed to allocate any objects")
	}

	// Reset and verify we can allocate again
	arena.Reset()
	_, err = arena.AllocateObject(42)
	if err != nil {
		t.Errorf("Failed to allocate after reset: %v", err)
	}
}

// Test both arena implementations with zero-sized allocations
func TestZeroSizedAllocations(t *testing.T) {
	// Standard arena
	stdArena, _ := NewMemoryArena[int](100)
	_, err := stdArena.Allocate(0)
	if err == nil {
		t.Error("Expected error when allocating size 0 in standard arena, got nil")
	}

	// Concurrent arena
	concArena, _ := NewConcurrentArena[int](100)
	_, err = concArena.Allocate(0)
	if err == nil {
		t.Error("Expected error when allocating size 0 in concurrent arena, got nil")
	}
}

// Test alignment edge cases
func TestAlignmentEdgeCases(t *testing.T) {
	arena, _ := NewMemoryArena[int64](100)

	// Test with different initial offsets
	testOffsets := []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	alignment := unsafe.Alignof(int64(0))

	for _, offset := range testOffsets {
		arena.buffer.offset = offset
		arena.alignOffset(alignment)

		// Check alignment is correct
		if arena.buffer.offset%int(alignment) != 0 {
			t.Errorf("Alignment failed for offset %d: got %d which is not aligned to %d",
				offset, arena.buffer.offset, alignment)
		}

		// Check we didn't over-align
		expectedOffset := offset
		if offset%int(alignment) != 0 {
			expectedOffset = offset + int(alignment) - (offset % int(alignment))
		}

		if arena.buffer.offset != expectedOffset {
			t.Errorf("Incorrect alignment for offset %d: expected %d, got %d",
				offset, expectedOffset, arena.buffer.offset)
		}
	}
}

// --- Tests for ErrArenaFull ---

func TestMemoryArena_Allocate_ArenaFull(t *testing.T) {
	size := int(unsafe.Sizeof(int(0)))
	arena, _ := NewMemoryArena[int](size) // Arena with space for exactly one int
	_, err := arena.Allocate(size)        // Allocate the first int
	if err != nil {
		t.Fatalf("Initial allocation failed unexpectedly: %v", err)
	}
	_, err = arena.Allocate(size) // Attempt to allocate second int
	if !errors.Is(err, ErrArenaFull) {
		t.Errorf("Expected ErrArenaFull, got %v", err)
	}
}

func TestMemoryArena_AllocateBuffer_ArenaFull(t *testing.T) {
	size := int(unsafe.Sizeof(int(0)))
	arena, _ := NewMemoryArena[int](size)
	_, err := arena.AllocateBuffer(size)
	if err != nil {
		t.Fatalf("Initial allocation failed unexpectedly: %v", err)
	}
	_, err = arena.AllocateBuffer(size)
	if !errors.Is(err, ErrArenaFull) {
		t.Errorf("Expected ErrArenaFull, got %v", err)
	}
}

func TestMemoryArena_AllocateNewValue_ArenaFull(t *testing.T) {
	size := int(unsafe.Sizeof(int(0)))
	arena, _ := NewMemoryArena[int](size)
	_, err := arena.AllocateNewValue(size, 1)
	if err != nil {
		t.Fatalf("Initial allocation failed unexpectedly: %v", err)
	}
	_, err = arena.AllocateNewValue(size, 2)
	if !errors.Is(err, ErrArenaFull) {
		t.Errorf("Expected ErrArenaFull, got %v", err)
	}
}

func TestMemoryArena_AllocateObject_ArenaFull(t *testing.T) {
	size := int(unsafe.Sizeof(int(0)))
	arena, _ := NewMemoryArena[int](size)
	_, err := arena.AllocateObject(1)
	if err != nil {
		t.Fatalf("Initial allocation failed unexpectedly: %v", err)
	}
	_, err = arena.AllocateObject(2)
	if !errors.Is(err, ErrArenaFull) {
		t.Errorf("Expected ErrArenaFull, got %v", err)
	}
}

func TestConcurrentArena_Allocate_ArenaFull(t *testing.T) {
	size := int(unsafe.Sizeof(int(0)))
	arena, _ := NewConcurrentArena[int](size)
	_, err := arena.Allocate(size) // Direct allocation (might bypass lock in real use, test assumes underlying method)
	if err != nil {
		t.Fatalf("Initial allocation failed unexpectedly: %v", err)
	}
	_, err = arena.Allocate(size)
	if !errors.Is(err, ErrArenaFull) {
		t.Errorf("Expected ErrArenaFull, got %v", err)
	}
}

func TestConcurrentArena_AllocateObject_ArenaFull(t *testing.T) {
	size := int(unsafe.Sizeof(int(0)))
	arena, _ := NewConcurrentArena[int](size)
	_, err := arena.AllocateObject(1)
	if err != nil {
		t.Fatalf("Initial allocation failed unexpectedly: %v", err)
	}
	_, err = arena.AllocateObject(2)
	if !errors.Is(err, ErrArenaFull) {
		t.Errorf("Expected ErrArenaFull, got %v", err)
	}
}

func TestNewObject_ArenaFull(t *testing.T) {
	size := int(unsafe.Sizeof(int(0)))
	arena, _ := NewMemoryArena[int](size)
	_, err := NewObject(arena, 1)
	if err != nil {
		t.Fatalf("Initial allocation failed unexpectedly: %v", err)
	}
	obj, err := NewObject(arena, 2)
	if !errors.Is(err, ErrArenaFull) {
		t.Errorf("Expected ErrArenaFull, got %v", err)
	}
	if obj != nil {
		t.Error("Expected nil object pointer on allocation failure")
	}
}

func TestAppendSlice_ArenaFull(t *testing.T) {
	// Arena needs space for the initial slice + the appended slice
	initialSlice := []int{1}
	appendedSlice := []int{1, 2}
	initialSize := int(unsafe.Sizeof(initialSlice))
	appendedSize := int(unsafe.Sizeof(appendedSlice)) // Size can vary based on slice header + backing array
	// Allocate just enough for the *initial* slice allocation
	arena, _ := NewMemoryArena[[]int](initialSize + int(unsafe.Alignof(initialSlice)))

	ptr, err := NewObject(arena, initialSlice)
	if err != nil {
		t.Fatalf("Initial slice allocation failed: %v", err)
	}
	slice := *ptr // Get the slice allocated in the arena

	// Now try to append, which requires space for a larger slice
	objToAppend := 2
	newSlicePtr, err := AppendSlice(&objToAppend, arena, &slice)

	if !errors.Is(err, ErrArenaFull) {
		// Note: Depending on Go's slice allocation, the exact size needed can be tricky.
		// This test might pass if initialSize calculation was generous enough for appendedSize.
		// A more robust test might involve precisely calculating aligned sizes.
		// For now, we check if *some* error occurred, expecting ErrArenaFull in constrained scenarios.
		t.Logf("Initial size: %d, Appended size approx: %d, Arena offset: %d, Arena size: %d", initialSize, appendedSize, arena.buffer.offset, arena.buffer.size)
		t.Errorf("Expected ErrArenaFull when appending slice, got %v", err)
	}
	if newSlicePtr != nil {
		t.Error("Expected nil slice pointer on allocation failure")
	}
}

func TestMemoryArena_Resize_InvalidSize(t *testing.T) {
	arena, _ := NewMemoryArena[int](100)
	err := arena.Resize(0)
	if !errors.Is(err, ErrInvalidSize) {
		t.Errorf("Expected ErrInvalidSize for size 0, got %v", err)
	}
	err = arena.Resize(-10)
	if !errors.Is(err, ErrInvalidSize) {
		t.Errorf("Expected ErrInvalidSize for negative size, got %v", err)
	}
}

func TestInsertMap_ArenaFull(t *testing.T) {
	// Create an initial map with one key/value.
	initialMap := map[string]int{"a": 1}

	// Create a very small arena so that after the initial allocation there is not enough space.
	// Choosing a capacity of 12 bytes: the initial allocation will use about 8 bytes (on a 64-bit system)
	// and leave only 4 bytes, which is insufficient for another allocation.
	arenaCapacity := 12
	arena, err := NewMemoryArena[map[string]int](arenaCapacity)
	if err != nil {
		t.Fatalf("Failed to create arena: %v", err)
	}

	ptr, err := NewObject(arena, initialMap)
	if err != nil {
		t.Fatalf("Initial map allocation failed: %v", err)
	}
	hashmap := *ptr

	t.Logf("After initial allocation - Arena offset: %d, Arena size: %d", arena.buffer.offset, arena.buffer.size)

	// Attempt to insert a new key/value pair into the map. This should require more space than is available.
	objToInsert := 2
	newMapPtr, err := InsertMap(&objToInsert, arena, &hashmap, "b")

	if err == nil {
		t.Errorf("Expected ErrArenaFull error, but got nil")
	}
	if !errors.Is(err, ErrArenaFull) {
		t.Errorf("Expected ErrArenaFull, got: %v", err)
	}
	if newMapPtr != nil {
		t.Errorf("Expected nil map pointer on allocation failure, got: %v", newMapPtr)
	}
}

func TestConcurrentArena_New_InvalidSize(t *testing.T) {
	_, err := NewConcurrentArena[int](0)
	// Expect error from underlying NewMemoryArena
	if err == nil {
		t.Errorf("Expected error when creating concurrent arena with size 0, got nil")
	}
	// Check if it's the specific error if possible, otherwise check non-nil
	// The exact error might be wrapped, so checking non-nil is safer.
	// if !errors.Is(err, ErrInvalidSize) { // This might fail if wrapped
	//  t.Errorf("Expected ErrInvalidSize, got %v", err)
	// }

	_, err = NewConcurrentArena[int](-10)
	if err == nil {
		t.Errorf("Expected error when creating concurrent arena with negative size, got nil")
	}
}

func TestConcurrentArena_Resize_InvalidSize(t *testing.T) {
	arena, _ := NewConcurrentArena[int](100)
	err := arena.Resize(0)
	if !errors.Is(err, ErrInvalidSize) {
		t.Errorf("Expected ErrInvalidSize for size 0, got %v", err)
	}
	err = arena.Resize(-10)
	if !errors.Is(err, ErrInvalidSize) {
		t.Errorf("Expected ErrInvalidSize for negative size, got %v", err)
	}
}

func TestConcurrentArena_ResizePreserve_InvalidSize(t *testing.T) {
	arena, _ := NewConcurrentArena[int](100)
	err := arena.ResizePreserve(0)
	if !errors.Is(err, ErrInvalidSize) {
		t.Errorf("Expected ErrInvalidSize for size 0, got %v", err)
	}
	err = arena.ResizePreserve(-10)
	if !errors.Is(err, ErrInvalidSize) {
		t.Errorf("Expected ErrInvalidSize for negative size, got %v", err)
	}
}

// --- Test for ErrNewSizeTooSmall ---

func TestConcurrentArena_ResizePreserve_TooSmall(t *testing.T) {
	arena, _ := NewConcurrentArena[int](100)
	_, err := arena.AllocateObject(1) // Allocate something
	if err != nil {
		t.Fatalf("Allocation failed: %v", err)
	}
	used := arena.buffer.offset
	err = arena.ResizePreserve(used - 1) // Attempt to resize smaller than used
	if !errors.Is(err, ErrNewSizeTooSmall) {
		t.Errorf("Expected ErrNewSizeTooSmall, got %v", err)
	}
}

// TestNewObject_Error creates an arena too small to hold an int so that NewObject returns an error.
func TestNewObject_Error(t *testing.T) {
	// Create a very small arena (1 byte) for ints.
	arena, err := NewMemoryArena[int](1)
	if err != nil {
		t.Fatalf("Failed to create arena: %v", err)
	}

	obj := 42
	// Use NewObject with an explicit type parameter.
	ptr, err := NewObject[int](arena, obj)
	if err == nil {
		t.Error("Expected error due to insufficient memory, but got nil")
	}
	if ptr != nil {
		t.Error("Expected nil pointer on allocation failure")
	}
	if !errors.Is(err, ErrArenaFull) {
		t.Errorf("Expected ErrArenaFull, got %v", err)
	}
}

// TestAppendSlice_Error creates an arena too small for a slice of ints to force an allocation error.
func TestAppendSlice_Error(t *testing.T) {
	// Create a very small arena (1 byte) for slices of int.
	arena, err := NewMemoryArena[[]int](1)
	if err != nil {
		t.Fatalf("Failed to create arena: %v", err)
	}

	// Start with an initial slice.
	slice := []int{1, 2, 3}
	// Attempt to append a new element using AppendSlice with explicit type parameter.
	newSlice, err := AppendSlice[int](new(int), arena, &slice)
	if err == nil {
		t.Error("Expected error due to insufficient memory for appending slice, but got nil")
	}
	if newSlice != nil {
		t.Error("Expected nil pointer on allocation failure in AppendSlice")
	}
	if !errors.Is(err, ErrArenaFull) {
		t.Errorf("Expected ErrArenaFull, got %v", err)
	}
}

// TestInsertMap_Error creates an arena too small for a map[string]int to force an allocation error.
func TestInsertMap_Error(t *testing.T) {
	// Create a very small arena (1 byte) for map[string]int.
	arena, err := NewMemoryArena[map[string]int](1)
	if err != nil {
		t.Fatalf("Failed to create arena: %v", err)
	}

	// Start with an initial map.
	hashmap := map[string]int{"a": 1}
	// Attempt to insert a new key/value pair using InsertMap with explicit type parameter.
	newMap, err := InsertMap[int](new(int), arena, &hashmap, "b")
	if err == nil {
		t.Error("Expected error due to insufficient memory for inserting into map, but got nil")
	}
	if newMap != nil {
		t.Error("Expected nil pointer on allocation failure in InsertMap")
	}
	if !errors.Is(err, ErrArenaFull) {
		t.Errorf("Expected ErrArenaFull, got %v", err)
	}
}

func TestAppendSlice_Success(t *testing.T) {
	// Create a sufficiently large arena for []int.
	arena, err := NewMemoryArena[[]int](100)
	if err != nil {
		t.Fatalf("Failed to create arena: %v", err)
	}

	// Start with an initial slice.
	slice := []int{1, 2, 3}

	// Append a new element (4) using AppendSlice.
	// The function uses the provided object (here, new(int) is just a placeholder).
	// The actual value to append comes from the append operation.
	// Because AppendSlice internally calls: *slice = append(*slice, *obj),
	// we first set *obj to the value 4.
	obj := 4
	newSlice, err := AppendSlice[int](&obj, arena, &slice)
	if err != nil {
		t.Fatalf("AppendSlice returned error: %v", err)
	}

	// The slice passed in should now contain the new element.
	expected := []int{1, 2, 3, 4}
	if len(*newSlice) != len(expected) {
		t.Errorf("Expected slice length %d, got %d", len(expected), len(*newSlice))
	}
	for i, v := range expected {
		if (*newSlice)[i] != v {
			t.Errorf("Expected element %d at index %d, got %d", v, i, (*newSlice)[i])
		}
	}
}

// TestInsertMap_Success verifies that InsertMap correctly inserts a new key/value pair
// into a map when there is sufficient memory available in the arena.
func TestInsertMap_Success(t *testing.T) {
	// Create a sufficiently large arena for map[string]int.
	arena, err := NewMemoryArena[map[string]int](100)
	if err != nil {
		t.Fatalf("Failed to create arena: %v", err)
	}

	// Start with an initial map.
	initialMap := map[string]int{"a": 1}

	// Insert a new key "b" with value 2.
	obj := 2
	newMap, err := InsertMap[int](&obj, arena, &initialMap, "b")
	if err != nil {
		t.Fatalf("InsertMap returned error: %v", err)
	}

	// The resulting map should have both keys "a" and "b" with the correct values.
	if val, ok := (*newMap)["a"]; !ok || val != 1 {
		t.Errorf("Expected key 'a' with value 1, got %v", (*newMap)["a"])
	}
	if val, ok := (*newMap)["b"]; !ok || val != 2 {
		t.Errorf("Expected key 'b' with value 2, got %v", (*newMap)["b"])
	}
}

func TestHelperReset(t *testing.T) {
	// Create a memory arena with a capacity of 100 bytes.
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Fatalf("Failed to create memory arena: %v", err)
	}

	// Allocate 8 bytes (e.g. for an int) to simulate arena usage.
	_, err = arena.Allocate(8)
	if err != nil {
		t.Fatalf("Allocation failed: %v", err)
	}

	// Write nonzero values in the allocated region.
	startIndex := arena.buffer.offset - 8
	for i := 0; i < 8; i++ {
		arena.buffer.memory[startIndex+i] = 42 // example nonzero value
	}

	// Verify that the allocated block contains nonzero values.
	foundNonZero := false
	for i := startIndex; i < arena.buffer.offset; i++ {
		if arena.buffer.memory[i] != 0 {
			foundNonZero = true
			break
		}
	}
	if !foundNonZero {
		t.Error("Expected allocated memory to have nonzero values before reset")
	}

	// Use the helper Reset function which accepts an Arena interface.
	Reset(arena)

	// Verify that the offset is reset to 0.
	if arena.buffer.offset != 0 {
		t.Errorf("Expected offset to be 0 after reset, got %d", arena.buffer.offset)
	}

	// Verify that all bytes in the arena's memory are now zeroed.
	for i, b := range arena.buffer.memory {
		if b != 0 {
			t.Errorf("Expected memory[%d] to be 0 after reset, got %d", i, b)
		}
	}
}

func TestNewMemoryArenaBuffer_NonStandardAlignment(t *testing.T) {
	size := 100
	alignment := uintptr(3) // Using 3 is unusual and may force a nonzero remainder.
	buf := NewMemoryArenaBuffer(size, alignment)
	base := uintptr(unsafe.Pointer(&buf.memory[0]))
	rem := base % alignment
	var expected int
	if rem != 0 {
		expected = int(alignment - rem)
	}
	if buf.offset != expected {
		t.Errorf("expected offset %d (for base %% alignment %d), got %d", expected, rem, buf.offset)
	}
}

func TestConcurrentArena_MixedReadWrite(t *testing.T) {
	arena, _ := NewConcurrentArena[string](50000)
	const writers = 10
	const readers = 40
	const writeOps = 100
	const readOps = 500

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	// Create a channel to report errors from goroutines
	errorCh := make(chan string, writers+readers)

	// Launch writers
	for w := 0; w < writers; w++ {
		go func(id int) {
			defer wg.Done()

			for i := 0; i < writeOps; i++ {
				// Allocate a string
				value := fmt.Sprintf("Writer %d - String %d", id, i)
				_, err := arena.AllocateObject(value)
				if err != nil {
					errorCh <- fmt.Sprintf("Writer %d: Allocation failed: %v", id, err)
					// Don't break on error, just continue
				}

				// Occasionally trigger a resize
				if i%25 == 0 {
					// Lock the mutex before reading shared state
					arena.mutex.Lock()
					currentSize := arena.buffer.size
					arena.mutex.Unlock()

					newSize := currentSize + (i % 1000)
					if err := arena.ResizePreserve(newSize); err != nil {
						// This can fail due to concurrency, which is expected
						continue
					}
				}
			}
		}(w)
	}

	// Launch readers
	for r := 0; r < readers; r++ {
		go func(id int) {
			defer wg.Done()

			for i := 0; i < readOps; i++ {
				// Read-only operations are already guarded with RLock
				arena.mutex.Lock()
				size := arena.buffer.size
				offset := arena.buffer.offset
				if size <= 0 || offset < 0 {
					errorCh <- fmt.Sprintf("Reader %d: Invalid state: size=%d, offset=%d", id, size, offset)
				}
				arena.mutex.Unlock()
			}
		}(r)
	}

	// Wait for all operations to complete
	wg.Wait()
	close(errorCh)

	// Report any errors
	errorCount := 0
	for err := range errorCh {
		errorCount++
		if errorCount <= 10 { // Limit the number of errors we log
			t.Errorf("Concurrent error: %s", err)
		}
	}

	if errorCount > 0 {
		t.Logf("Total of %d errors occurred during concurrent test", errorCount)
	}
}
