package memoryArena

import (
	"testing"
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

func BenchmarkConcurrentArena_AllocateObject(b *testing.B) {
	arena, err := NewConcurrentArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	obj := 5
	for i := 0; i < b.N; i++ {
		_, err = arena.AllocateObject(obj)
		if err != nil {
			b.Errorf("Error: %v", err)
		}
	}
}

func BenchmarkConcurrentArena_AllocateNewValue(b *testing.B) {
	arena, err := NewConcurrentArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	obj := 5
	for i := 0; i < b.N; i++ {
		_, err = arena.AllocateNewValue(10, obj)
		if err != nil {
			b.Errorf("Error: %v", err)
		}
	}
}

func BenchmarkConcurrentArena_Allocate(b *testing.B) {
	arena, err := NewConcurrentArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		_, err = arena.Allocate(10)
		if err != nil {
			b.Errorf("Error: %v", err)
		}
	}
}

func BenchmarkConcurrentArena_ResizePreserve(b *testing.B) {
	arena, err := NewConcurrentArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		err = arena.ResizePreserve(100)
		if err != nil {
			b.Errorf("Error: %v", err)
		}
	}
}

func BenchmarkConcurrentArena_Resize(b *testing.B) {
	arena, err := NewConcurrentArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		err = arena.Resize(100)
		if err != nil {
			b.Errorf("Error: %v", err)
		}
	}
}

func BenchmarkConcurrentArena_Free(b *testing.B) {
	arena, err := NewConcurrentArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		arena.Free()
	}
}

func BenchmarkConcurrentArena_GetResult(b *testing.B) {
	arena, err := NewConcurrentArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		arena.GetResult()
	}
}

func BenchmarkConcurrentArena_Reset(b *testing.B) {
	arena, err := NewConcurrentArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		arena.Reset()
	}
}
