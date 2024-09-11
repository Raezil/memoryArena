package memoryArena

import "testing"

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

func TestNewObject(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	obj := 5
	_, err = NewObject(arena, obj)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

func TestAppendSlice(t *testing.T) {
	arena, err := NewMemoryArena[[]int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	obj := 5
	slice := []int{1, 2, 3}
	_, err = AppendSlice(&obj, arena, &slice)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

func TestSetNewValue(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	obj := 5
	ptr, err := arena.Allocate(10)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	ptr, _ = SetNewValue(&ptr, obj)
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

func BenchmarkMemoryArena_AllocateObject(b *testing.B) {
	arena, err := NewMemoryArena[int](100)
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

func TestInsertMap(t *testing.T) {
	arena, err := NewMemoryArena[map[string]int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	obj := 4
	slice := map[string]int{"1": 1, "2": 2, "3": 3}
	_, err = InsertMap(&obj, arena, &slice, "4")
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
