package memoryArena

import (
	"testing"
)

func TestConcurrentArena(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	concurrentArena := NewConcurrentArena(*arena)
	ptr, err := concurrentArena.Allocate(10)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if ptr == nil {
		t.Errorf("Error: ptr is nil")
	}

}
func TestConcurrentArena_AllocateObject(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	concurrentArena := NewConcurrentArena(*arena)
	obj := 5
	_, err = concurrentArena.AllocateObject(obj)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

func BenchmarkConcurrentArena_AllocateObject(b *testing.B) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	concurrentArena := NewConcurrentArena(*arena)
	obj := 5
	for i := 0; i < b.N; i++ {
		_, err = concurrentArena.AllocateObject(obj)
		if err != nil {
			b.Errorf("Error: %v", err)
		}
	}
}

func TestConcurrentArena_Free(t *testing.T) {
	memoryarena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	arena := NewConcurrentArena[int](*memoryarena)
	arena.Free()

	for i := range arena.buffer.memory {
		if arena.buffer.memory[i] != 0 {
			t.Errorf("Error: memory is not freed")
		}
	}
}
