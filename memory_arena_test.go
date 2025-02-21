package memoryArena

import (
	"testing"
)

func TestMemoryArena_Allocate(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
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

func TestMemoryArena_ResetAndFree(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Fatalf("Error creating arena: %v", err)
	}
	// Allocate some memory.
	_, err = arena.Allocate(10)
	if err != nil {
		t.Fatalf("Error allocating memory: %v", err)
	}
	arena.Reset()
	for i, bVal := range arena.buffer.memory {
		if bVal != 0 {
			t.Errorf("Memory not reset at index %d: got %d", i, bVal)
		}
	}
	// Allocate again and then free.
	_, err = arena.Allocate(10)
	if err != nil {
		t.Fatalf("Error allocating memory: %v", err)
	}
	arena.Free()
	for i, bVal := range arena.buffer.memory {
		if bVal != 0 {
			t.Errorf("Memory not freed at index %d: got %d", i, bVal)
		}
	}
}

func TestMemoryArena_ResizePreserve(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Fatalf("Error creating arena: %v", err)
	}
	obj := 5
	p, err := NewObject(arena, obj)
	if err != nil {
		t.Fatalf("Error allocating object: %v", err)
	}
	err = arena.ResizePreserve(200)
	if err != nil {
		t.Fatalf("Error resizing preserve: %v", err)
	}
	if arena.buffer.size != 200 {
		t.Errorf("Expected arena size 200, got %d", arena.buffer.size)
	}
	if *p != obj {
		t.Errorf("Expected object %d, got %d", obj, *p)
	}
}

func TestMemoryArena_Resize(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Fatalf("Error creating arena: %v", err)
	}
	err = arena.Resize(200)
	if err != nil {
		t.Fatalf("Error resizing arena: %v", err)
	}
	if arena.buffer.size != 200 {
		t.Errorf("Expected arena size 200, got %d", arena.buffer.size)
	}
}

func TestMemoryArena_AllocateNewValue(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Fatalf("Error creating arena: %v", err)
	}
	obj := 5
	p, err := arena.AllocateNewValue(obj)
	if err != nil {
		t.Fatalf("Error allocating new value: %v", err)
	}
	if p == nil {
		t.Fatalf("Allocated pointer is nil")
	}
	if *p != obj {
		t.Errorf("Expected value %d, got %d", obj, *p)
	}
}

func TestMemoryArena_GetResult(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Fatalf("Error creating arena: %v", err)
	}
	// Change offset by allocating memory.
	_, err = arena.Allocate(10)
	if err != nil {
		t.Fatalf("Error allocating memory: %v", err)
	}
	result := arena.GetResult()
	if result == nil {
		t.Fatalf("GetResult returned nil")
	}
}

func BenchmarkMemoryArena_AllocateNewValue(b *testing.B) {
	arena, err := NewMemoryArena[int](1000)
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

func BenchmarkMemoryArena_Allocate(b *testing.B) {
	arena, err := NewMemoryArena[int](1000)
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

func BenchmarkMemoryArena_Reset(b *testing.B) {
	arena, err := NewMemoryArena[int](1000)
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

func BenchmarkMemoryArena_Free(b *testing.B) {
	arena, err := NewMemoryArena[int](1000)
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

func BenchmarkMemoryArena_GetResult(b *testing.B) {
	arena, err := NewMemoryArena[int](1000)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = arena.GetResult()
	}
}

func BenchmarkMemoryArena_ResizePreserve(b *testing.B) {
	arena, err := NewMemoryArena[int](1000)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
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
		arena.Reset()
	}
}

func BenchmarkMemoryArena_Resize(b *testing.B) {
	arena, err := NewMemoryArena[int](1000)
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
