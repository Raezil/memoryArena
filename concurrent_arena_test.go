package memoryArena

import (
	"testing"
)

func TestConcurrentArena_Allocate(t *testing.T) {
	arena, err := NewConcurrentArena[int](100)
	if err != nil {
		t.Fatalf("Error creating arena: %v", err)
	}
	ptr, err := arena.Allocate(10)
	if err != nil {
		t.Fatalf("Error allocating memory: %v", err)
	}
	if ptr == nil {
		t.Fatalf("Allocated pointer is nil")
	}
}

func TestConcurrentArena_AllocateNewValue(t *testing.T) {
	arena, err := NewConcurrentArena[int](100)
	if err != nil {
		t.Fatalf("Error creating arena: %v", err)
	}
	obj := 5
	p, err := arena.AllocateNewValue(obj)
	if err != nil {
		t.Fatalf("Error allocating new value: %v", err)
	}
	if p == nil {
		t.Fatalf("Allocated new value pointer is nil")
	}
	if *p != obj {
		t.Fatalf("Expected %d, got %d", obj, *p)
	}
}

func TestConcurrentArena_Free(t *testing.T) {
	arena, err := NewConcurrentArena[int](100)
	if err != nil {
		t.Fatalf("Error creating arena: %v", err)
	}
	// allocate some memory first
	_, err = arena.Allocate(10)
	if err != nil {
		t.Fatalf("Error allocating memory: %v", err)
	}
	arena.Free()
	// verify that all bytes in the arena are zeroed
	for i, b := range arena.buffer.memory {
		if b != 0 {
			t.Errorf("Memory not freed at index %d: got %d", i, b)
		}
	}
}

func BenchmarkConcurrentArena_AllocateNewValue(b *testing.B) {
	arena, err := NewConcurrentArena[int](1000)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	obj := 5
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = arena.AllocateNewValue(obj)
		if err != nil {
			b.Fatalf("Error allocating new value: %v", err)
		}
	}
}

func BenchmarkConcurrentArena_Allocate(b *testing.B) {
	arena, err := NewConcurrentArena[int](1000)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = arena.Allocate(10)
		if err != nil {
			b.Fatalf("Error allocating memory: %v", err)
		}
	}
}

func BenchmarkConcurrentArena_ResizePreserve(b *testing.B) {
	arena, err := NewConcurrentArena[int](1000)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	// allocate some memory to ensure used space exists
	_, err = arena.Allocate(50)
	if err != nil {
		b.Fatalf("Error allocating memory: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = arena.ResizePreserve(2000)
		if err != nil {
			b.Fatalf("Error resizing preserve: %v", err)
		}
		// reset for the next iteration
		arena.Reset()
	}
}

func BenchmarkConcurrentArena_Resize(b *testing.B) {
	arena, err := NewConcurrentArena[int](1000)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = arena.Resize(1000)
		if err != nil {
			b.Fatalf("Error resizing arena: %v", err)
		}
	}
}

func BenchmarkConcurrentArena_Free(b *testing.B) {
	arena, err := NewConcurrentArena[int](1000)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	_, err = arena.Allocate(50)
	if err != nil {
		b.Fatalf("Error allocating memory: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.Free()
	}
}

func BenchmarkConcurrentArena_GetResult(b *testing.B) {
	arena, err := NewConcurrentArena[int](1000)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = arena.GetResult()
	}
}

func BenchmarkConcurrentArena_Reset(b *testing.B) {
	arena, err := NewConcurrentArena[int](1000)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	_, err = arena.Allocate(50)
	if err != nil {
		b.Fatalf("Error allocating memory: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.Reset()
	}
}
