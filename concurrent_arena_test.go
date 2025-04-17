package memoryArena

import (
	"sync"
	"testing"
	"unsafe"
)

func TestConcurrentArena(t *testing.T) {
	arena, err := NewConcurrentArena[int](100)
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
func TestConcurrentArena_AllocateObject(t *testing.T) {
	arena, err := NewConcurrentArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	obj := 5
	_, err = arena.AllocateObject(obj)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

func TestConcurrentArena_Free(t *testing.T) {
	arena, err := NewConcurrentArena[int](100)
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

// Test that ConcurrentArena AllocateObject returns an error when given a wrong type.
func TestConcurrentArena_AllocateObject_WrongType(t *testing.T) {
	arena, err := NewConcurrentArena[int](10)
	if err != nil {
		t.Fatalf("Failed to create concurrent arena: %v", err)
	}
	_, err = arena.AllocateObject("wrong type")
	if err == nil {
		t.Error("Expected error when allocating object with wrong type in concurrent arena, got nil")
	}
}

// Test that ConcurrentArena Reset works correctly when called concurrently.
func TestConcurrentArena_Reset_Concurrent(t *testing.T) {
	arena, err := NewConcurrentArena[int](10)
	if err != nil {
		t.Fatalf("Failed to create concurrent arena: %v", err)
	}
	// Allocate some data.
	_, err = arena.Allocate(10)
	if err != nil {
		t.Fatalf("Allocation failed: %v", err)
	}
	done := make(chan bool)
	go func() {
		arena.Reset()
		done <- true
	}()
	<-done
	if arena.buffer.offset != 0 {
		t.Error("Expected offset 0 after concurrent reset")
	}
}

// TESTS

const benchmarkConcurrentArenaSize = 1 << 20 // 1 MB - adjust as needed

func BenchmarkConcurrentArena_AllocateObject(b *testing.B) {
	arena, err := NewConcurrentArena[int](benchmarkConcurrentArenaSize) // Increased size
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

func BenchmarkConcurrentArena_AllocateNewValue(b *testing.B) {
	arena, err := NewConcurrentArena[int](benchmarkConcurrentArenaSize) // Increased size
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	obj := 5
	size := int(unsafe.Sizeof(obj)) // Calculate size
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.Reset() // Reset for each allocation benchmark iteration
		// Assuming AllocateNewValue exists on the non-concurrent arena accessed via embedding
		_, err = arena.MemoryArena.AllocateNewValue(size, obj) // Call embedded method if needed
		if err != nil {
			b.Fatalf("Iteration %d: AllocateNewValue failed: %v", i, err) // Use Fatalf
		}
	}
}

func BenchmarkConcurrentArena_Allocate(b *testing.B) {
	arena, err := NewConcurrentArena[int](benchmarkConcurrentArenaSize) // Increased size
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

// Benchmarking Resize/ResizePreserve/Free/GetResult/Reset for ConcurrentArena
// follows similar logic to MemoryArena benchmarks. Ensure Fatalf and consider setup/reset logic.

func BenchmarkConcurrentArena_ResizePreserve(b *testing.B) {
	initialSize := 1024
	resizeTo := 2048
	arena, err := NewConcurrentArena[int](initialSize)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = arena.ResizePreserve(resizeTo)
		if err != nil {
			// Can fail if new size is too small, check error if needed
			// If just benchmarking call speed, maybe ignore specific errors?
			// But Fatalf is safer if *any* error invalidates the timing.
			b.Fatalf("ResizePreserve failed: %v", err)
		}
		// Reset size for consistency if desired
		arena.Resize(initialSize) // Use underlying Resize for simplicity here?
	}
}

func BenchmarkConcurrentArena_Resize(b *testing.B) {
	initialSize := 1024
	resizeTo := 2048
	arena, err := NewConcurrentArena[int](initialSize)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = arena.Resize(resizeTo)
		if err != nil {
			b.Fatalf("Resize failed: %v", err)
		}
		// Reset size for consistency if desired
		arena.Resize(initialSize)
	}
}

func BenchmarkConcurrentArena_Free(b *testing.B) {
	arena, err := NewConcurrentArena[int](1024) // Moderate size
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.Free() // Benchmarking Free operation
	}
}

func BenchmarkConcurrentArena_GetResult(b *testing.B) {
	arena, err := NewConcurrentArena[int](1024) // Moderate size
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = arena.GetResult() // Benchmarking GetResult operation
	}
}

func BenchmarkConcurrentArena_Reset(b *testing.B) {
	arena, err := NewConcurrentArena[int](1024) // Moderate size
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.Reset() // Benchmarking Reset operation
	}
}

// Test that ConcurrentArena Reset clears memory when called concurrently.
func TestConcurrentArena_Reset_ConcurrentMemory(t *testing.T) {
	arena, err := NewConcurrentArena[int](160)
	if err != nil {
		t.Fatalf("Failed to create concurrent arena: %v", err)
	}
	// Allocate and initialize memory
	for i := 0; i < 20; i++ {
		_, err := arena.AllocateObject(i)
		if err != nil {
			t.Fatalf("Allocation failed at iteration %d: %v", i, err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			arena.Reset()
			wg.Done()
		}()
	}
	wg.Wait()

	// After reset, memory should be zeroed
	for idx, val := range arena.buffer.memory {
		if val != 0 {
			t.Errorf("Error: memory not cleared at index %d, got %v", idx, val)
		}
	}
}

// Test that ConcurrentArena Free works correctly when called concurrently.
func TestConcurrentArena_Free_Concurrent(t *testing.T) {
	arena, err := NewConcurrentArena[int](1000)
	if err != nil {
		t.Fatalf("Failed to create concurrent arena: %v", err)
	}
	// Allocate and initialize memory
	for i := 0; i < 50; i++ {
		_, err := arena.AllocateObject(i)
		if err != nil {
			t.Fatalf("Allocation failed at iteration %d: %v", i, err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			arena.Free()
			wg.Done()
		}()
	}
	wg.Wait()

	// After free, memory should be zeroed
	for idx, val := range arena.buffer.memory {
		if val != 0 {
			t.Errorf("Error: memory not freed at index %d, got %v", idx, val)
		}
	}
}

// --- New concurrent tests ---

// Test concurrent allocations to ensure thread-safe Allocate implementation.
func TestConcurrentArena_Allocate_Concurrent(t *testing.T) {
	arena, err := NewConcurrentArena[int](1000)
	if err != nil {
		t.Fatalf("Failed to create concurrent arena: %v", err)
	}
	var wg sync.WaitGroup
	const count = 100
	results := make([]unsafe.Pointer, count)
	wg.Add(count)
	for i := 0; i < count; i++ {
		i := i
		go func() {
			ptr, err := arena.Allocate(1)
			if err != nil {
				t.Errorf("Concurrent Allocate failed: %v", err)
			}
			results[i] = ptr
			wg.Done()
		}()
	}
	wg.Wait()
	// Check non-nil and uniqueness
	seen := make(map[uintptr]bool)
	for i, ptr := range results {
		if ptr == nil {
			t.Errorf("ptr[%d] is nil", i)
		}
		addr := uintptr(ptr)
		if seen[addr] {
			t.Errorf("Duplicate pointer at iteration %d: %v", i, ptr)
		}
		seen[addr] = true
	}
}

// Test concurrent object allocations to ensure AllocateObject is thread-safe and preserves values.
func TestConcurrentArena_AllocateObject_Concurrent(t *testing.T) {
	arena, err := NewConcurrentArena[int](1000)
	if err != nil {
		t.Fatalf("Failed to create concurrent arena: %v", err)
	}
	values := []int{10, 20, 30, 40, 50}
	var wg sync.WaitGroup
	count := len(values)
	results := make([]int, count)
	wg.Add(count)
	for idx, v := range values {
		idx, v := idx, v
		go func() {
			ptr, err := arena.AllocateObject(v)
			if err != nil {
				t.Errorf("Concurrent AllocateObject failed: %v", err)
			}
			results[idx] = *(*int)(ptr)
			wg.Done()
		}()
	}
	wg.Wait()
	for i, v := range values {
		if results[i] != v {
			t.Errorf("Value mismatch at index %d: expected %d, got %d", i, v, results[i])
		}
	}
}

// Test concurrent resize operations to ensure thread-safe Resize implementation.
func TestConcurrentArena_Resize_Concurrent(t *testing.T) {
	arena, err := NewConcurrentArena[int](10)
	if err != nil {
		t.Fatalf("Failed to create concurrent arena: %v", err)
	}
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			if err := arena.Resize(20); err != nil {
				t.Errorf("Concurrent Resize failed: %v", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if len(arena.buffer.memory) != 20 {
		t.Errorf("Expected capacity 20 after concurrent Resize, got %d", len(arena.buffer.memory))
	}
}

// Test concurrent preserve-resize operations to ensure thread-safe ResizePreserve and data integrity.
func TestConcurrentArena_ResizePreserve_Concurrent(t *testing.T) {
	arena, err := NewConcurrentArena[int](10)
	if err != nil {
		t.Fatalf("Failed to create concurrent arena: %v", err)
	}
	// Allocate a value to preserve
	ptrInit, err := arena.AllocateObject(99)
	if err != nil {
		t.Fatalf("Initial allocation failed: %v", err)
	}
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			if err := arena.ResizePreserve(20); err != nil {
				t.Errorf("Concurrent ResizePreserve failed: %v", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if len(arena.buffer.memory) != 20 {
		t.Errorf("Expected capacity 20 after concurrent ResizePreserve, got %d", len(arena.buffer.memory))
	}
	// Check preserved initial value
	val := *(*int)(ptrInit)
	if val != 99 {
		t.Errorf("Expected preserved value 99 after ResizePreserve, got %d", val)
	}
}
