package memoryArena

import (
	"testing"
	"unsafe"
)

func TestMemoryArena(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	ptr, err := arena.Allocate(10)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if ptr == nil {
		t.Errorf("Error: ptr is nil")
	}

}

func TestMemoryArena_Reset(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	arena.Reset()

	for i := range arena.buffer.memory {
		if arena.buffer.memory[i] != 0 {
			t.Errorf("Error: memory is not reset")
		}
	}
}

func TestMemoryArena_Free(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	arena.Free()

	for i := range arena.buffer.memory {
		if arena.buffer.memory[i] != 0 {
			t.Errorf("Error: memory is not freed")
		}
	}
}

func TestMemoryArena_notEnoughSpace(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if arena.notEnoughSpace(100) {
		t.Errorf("Error: out of bounds")
	}
}

func TestMemoryArena_nextOffset(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if arena.nextOffset(10) != 10 {
		t.Errorf("Error: used capacity is not correct")
	}
}

func TestMemoryArena_alignOffset(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	arena.alignOffset(8)
	if arena.buffer.offset != 0 {
		t.Errorf("Error: offset is not aligned")
	}
}

func TestMemoryArenaBuffer_NewMemoryArenaBuffer(t *testing.T) {
	arena := NewMemoryArenaBuffer(100, 1)
	if arena.size != 100 {
		t.Errorf("Error: size is not correct")
	}
	if arena.offset != 0 {
		t.Errorf("Error: offset is not correct")
	}
}

func TestMemoryArenaResizePreserve(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	num, err := NewObject(arena, 5)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	arena.ResizePreserve(200)
	if arena.buffer.size != 200 {
		t.Errorf("Error: size is not preserved")
	}
	if *num != 5 {
		t.Errorf("Error: object is not preserved")
	}

}

func TestMemoryArenaResize(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	arena.Resize(200)
	if arena.buffer.size != 200 {
		t.Errorf("Error: size is not resized")
	}

}

func TestMemoryArenaAllocate(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	ptr, err := arena.Allocate(10)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if ptr == nil {
		t.Errorf("Error: ptr is nil")
	}

}

func TestMemoryArenaAllocateBuffer(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	ptr, err := arena.AllocateBuffer(10)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if ptr == nil {
		t.Errorf("Error: ptr is nil")
	}

}

func TestMemoryArena_AllocateNewValue(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	obj := 5
	ptr, err := arena.AllocateNewValue(10, obj)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if ptr == nil {
		t.Errorf("Error: ptr is nil")
	}
}

func TestMemoryArena_GetResult(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	ptr := arena.GetResult()
	if ptr == nil {
		t.Errorf("Error: ptr is nil")
	}
}

func TestMemoryArena_AllocateObject(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	obj := 5
	_, err = arena.AllocateObject(obj)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

// TestMemoryArenaBuffer_AlignmentCheck verifies that the effective base of the arena buffer is aligned.
func TestMemoryArenaBuffer_AlignmentCheck(t *testing.T) {
	// Test for a couple of alignment values.
	testCases := []struct {
		name      string
		alignment uintptr
	}{
		{"Alignment1", 1},
		{"Alignment8", 8},
		{"Alignment16", 16},
	}

	for _, tc := range testCases {
		// Create a new arena buffer with the given alignment.
		buffer := NewMemoryArenaBuffer(100, tc.alignment)
		basePtr := uintptr(unsafe.Pointer(&buffer.memory[0]))
		effectivePtr := basePtr + uintptr(buffer.offset)
		if effectivePtr%tc.alignment != 0 {
			t.Errorf("%s: effective pointer %#x is not aligned to %d bytes",
				tc.name, effectivePtr, tc.alignment)
		} else {
			t.Logf("%s: effective pointer %#x is correctly aligned to %d bytes",
				tc.name, effectivePtr, tc.alignment)
		}
	}
}

// TestMemoryArena_AllocationAlignment ensures that allocations return pointers that obey the type’s alignment.
func TestMemoryArena_AllocationAlignment(t *testing.T) {
	// Use int64 for which alignment is typically 8 bytes on 64-bit systems.
	arena, err := NewMemoryArena[int](10)
	if err != nil {
		t.Fatalf("Failed to create arena: %v", err)
	}

	ptr, err := arena.Allocate(8) // Allocate space for an int64.
	if err != nil {
		t.Fatalf("Allocation failed: %v", err)
	}

	effective := uintptr(ptr)
	requiredAlignment := uintptr(unsafe.Alignof(int64(0)))
	if effective%requiredAlignment != 0 {
		t.Errorf("Allocated pointer %#x is not aligned to %d bytes", effective, requiredAlignment)
	} else {
		t.Logf("Allocated pointer %#x is correctly aligned to %d bytes", effective, requiredAlignment)
	}
}

// TestMemoryArena_AllocateObjectAlignment checks that the pointer returned by AllocateObject
// is also aligned according to the type requirements.
func TestMemoryArena_AllocateObjectAlignment(t *testing.T) {
	type MyStruct struct {
		A int64
		B int32
	}
	arena, err := NewMemoryArena[MyStruct](100)
	if err != nil {
		t.Fatalf("Failed to create arena: %v", err)
	}

	obj := MyStruct{A: 42, B: 7}
	ptr, err := arena.AllocateObject(obj)
	if err != nil {
		t.Fatalf("AllocateObject failed: %v", err)
	}

	effective := uintptr(ptr)
	requiredAlignment := uintptr(unsafe.Alignof(MyStruct{}))
	if effective%requiredAlignment != 0 {
		t.Errorf("Object pointer %#x is not aligned to %d bytes", effective, requiredAlignment)
	} else {
		t.Logf("Object pointer %#x is correctly aligned to %d bytes", effective, requiredAlignment)
	}
}

// TestNewMemoryArenaBuffer_AlignmentOne verifies that when using an alignment of 1,
// no offset is needed.
func TestNewMemoryArenaBuffer_AlignmentOne(t *testing.T) {
	buf := NewMemoryArenaBuffer(100, 1)
	if buf.size != 100 {
		t.Errorf("expected size 100, got %d", buf.size)
	}
	if buf.offset != 0 {
		t.Errorf("expected offset 0 for alignment 1, got %d", buf.offset)
	}
	// Effective pointer check
	base := uintptr(unsafe.Pointer(&buf.memory[0]))
	effective := base + uintptr(buf.offset)
	if effective%1 != 0 {
		t.Errorf("effective pointer %#x is not aligned to 1 byte", effective)
	}
}

// TestNewMemoryArenaBuffer_LargeAlignment forces a misalignment branch.
// Using a large alignment (e.g. 4096) makes it very unlikely that the base is already aligned.
func TestNewMemoryArenaBuffer_LargeAlignment(t *testing.T) {
	const alignment = 4096
	buf := NewMemoryArenaBuffer(100, alignment)
	base := uintptr(unsafe.Pointer(&buf.memory[0]))
	remainder := base % alignment
	var expectedOffset int
	if remainder == 0 {
		// In the unlikely event that the base is aligned.
		expectedOffset = 0
	} else {
		expectedOffset = int(alignment - remainder)
	}
	if buf.offset != expectedOffset {
		t.Errorf("expected offset %d (base=0x%x, remainder=%d), got %d",
			expectedOffset, base, remainder, buf.offset)
	}
	// Ensure the effective pointer is aligned
	effective := base + uintptr(buf.offset)
	if effective%alignment != 0 {
		t.Errorf("effective pointer %#x is not aligned to %d bytes", effective, alignment)
	}
}

// TestNewMemoryArenaBuffer_AlternateAlignment tests a typical alignment (e.g. 8)
// and verifies that the computed offset is correct.
func TestNewMemoryArenaBuffer_AlternateAlignment(t *testing.T) {
	const alignment = 8
	buf := NewMemoryArenaBuffer(100, alignment)
	base := uintptr(unsafe.Pointer(&buf.memory[0]))
	remainder := base % alignment
	var expectedOffset int
	if remainder == 0 {
		expectedOffset = 0
	} else {
		expectedOffset = int(alignment - remainder)
	}
	if buf.offset != expectedOffset {
		t.Errorf("expected offset %d for alignment %d (base=0x%x, remainder=%d), got %d",
			expectedOffset, alignment, base, remainder, buf.offset)
	}
	effective := base + uintptr(buf.offset)
	if effective%alignment != 0 {
		t.Errorf("effective pointer %#x is not aligned to %d bytes", effective, alignment)
	}
}

const benchmarkArenaSize = 1 << 20 // 1 MB - adjust as needed

func BenchmarkMemoryArena_AllocateObject(b *testing.B) {
	arena, err := NewMemoryArena[int](benchmarkArenaSize) // Increased size
	if err != nil {
		b.Fatalf("Error creating arena: %v", err) // Use Fatalf for setup errors
	}
	obj := 5
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.Reset() // Reset for each allocation benchmark iteration
		_, err = arena.AllocateObject(obj)
		if err != nil {
			b.Fatalf("Iteration %d: AllocateObject failed: %v", i, err) // Use Fatalf
		}
	}
}

func BenchmarkMemoryArena_AllocateNewValue(b *testing.B) {
	arena, err := NewMemoryArena[int](benchmarkArenaSize) // Increased size
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	obj := 5
	size := int(unsafe.Sizeof(obj))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.Reset() // Reset for each allocation benchmark iteration
		_, err = arena.AllocateNewValue(size, obj)
		if err != nil {
			b.Fatalf("Iteration %d: AllocateNewValue failed: %v", i, err) // Use Fatalf
		}
	}
}

func BenchmarkMemoryArena_AllocateBuffer(b *testing.B) {
	arena, err := NewMemoryArena[int](benchmarkArenaSize) // Increased size
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	allocSize := 10
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.Reset() // Reset for each allocation benchmark iteration
		_, err = arena.AllocateBuffer(allocSize)
		if err != nil {
			b.Fatalf("Iteration %d: AllocateBuffer failed: %v", i, err) // Use Fatalf
		}
	}
}

func BenchmarkMemoryArena_Allocate(b *testing.B) {
	arena, err := NewMemoryArena[int](benchmarkArenaSize) // Increased size
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	allocSize := 10
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.Reset() // Reset for each allocation benchmark iteration
		_, err = arena.Allocate(allocSize)
		if err != nil {
			b.Fatalf("Iteration %d: Allocate failed: %v", i, err) // Use Fatalf
		}
	}
}

func BenchmarkMemoryArena_Reset(b *testing.B) {
	// Benchmarking Reset itself doesn't require a huge arena unless setup involves filling it.
	arena, err := NewMemoryArena[int](1024) // Moderate size is likely fine
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	// Optional: Allocate some data before timing if you want Reset to clear something
	// _, _ = arena.Allocate(512)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.Reset() // This is the operation being benchmarked
	}
}

func BenchmarkMemoryArena_Free(b *testing.B) {
	// Benchmarking Free itself
	arena, err := NewMemoryArena[int](1024) // Moderate size
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	// Optional: Allocate some data before timing if you want Free to clear something
	// _, _ = arena.Allocate(512)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.Free() // This is the operation being benchmarked
	}
}

func BenchmarkMemoryArena_GetResult(b *testing.B) {
	arena, err := NewMemoryArena[int](1024) // Moderate size
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = arena.GetResult() // This is the operation being benchmarked
	}
}

// Note: Benchmarking Resize/ResizePreserve without allocations might not be very informative.
// A more realistic benchmark might allocate data, then resize.
// Keeping the original structure for now.

func BenchmarkMemoryArena_ResizePreserve(b *testing.B) {
	// Initial size should be large enough if allocations happen before resize.
	initialSize := 1024
	resizeTo := 2048
	arena, err := NewMemoryArena[int](initialSize)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	// Optional: Allocate some data here before timing
	// _, _ = arena.Allocate(initialSize / 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset state if needed for consistent resize benchmark
		// arena.Reset()
		// arena.Allocate(initialSize / 2) // Re-allocate if testing resize with data
		err = arena.ResizePreserve(resizeTo)
		// Reset back to initial size for next iteration if needed?
		// Or just keep resizing? Depends on what's being measured.
		// If just the call:
		if err != nil {
			b.Fatalf("ResizePreserve failed: %v", err) // Use Fatalf
		}
		// If testing repeated resize on same (growing) arena, might need larger initial/target sizes
		// Or reset the arena size back if possible/meaningful
		arena.Resize(initialSize) // Reset size for next iteration consistency (if desired)

	}
}

func BenchmarkMemoryArena_Resize(b *testing.B) {
	initialSize := 1024
	resizeTo := 2048
	arena, err := NewMemoryArena[int](initialSize)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = arena.Resize(resizeTo)
		if err != nil {
			b.Fatalf("Resize failed: %v", err) // Use Fatalf
		}
		// Reset size for next iteration consistency (if desired)
		arena.Resize(initialSize)
	}
}
