package memoryArena

import (
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
