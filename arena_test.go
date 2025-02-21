package memoryArena

import (
	"testing"
)

func TestNewObject(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Fatalf("Error creating arena: %v", err)
	}
	obj := 5
	p, err := NewObject(arena, obj)
	if err != nil {
		t.Fatalf("Error creating new object: %v", err)
	}
	if p == nil || *p != obj {
		t.Fatalf("Expected object %d, got %v", obj, p)
	}
}

func TestAppendSlice(t *testing.T) {
	// Create an arena for slices of int.
	arena, err := NewMemoryArena[[]int](100)
	if err != nil {
		t.Fatalf("Error creating arena: %v", err)
	}
	initialSlice := []int{1, 2, 3}
	element := 5
	newSlicePtr, err := AppendSlice(arena, element, initialSlice)
	if err != nil {
		t.Fatalf("Error appending slice: %v", err)
	}
	if newSlicePtr == nil {
		t.Fatalf("New slice pointer is nil")
	}
	if len(*newSlicePtr) != len(initialSlice)+1 {
		t.Errorf("Expected slice length %d, got %d", len(initialSlice)+1, len(*newSlicePtr))
	}
}

func BenchmarkAppendSlice(b *testing.B) {
	arena, err := NewMemoryArena[[]int](1000)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	initialSlice := []int{1, 2, 3}
	element := 5
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = AppendSlice(arena, element, initialSlice)
		if err != nil {
			b.Fatalf("Error appending slice: %v", err)
		}
	}
}

func TestInsertMap(t *testing.T) {
	// Create an arena for map[string]int.
	arena, err := NewMemoryArena[map[string]int](100)
	if err != nil {
		t.Fatalf("Error creating arena: %v", err)
	}
	initialMap := map[string]int{"a": 1, "b": 2}
	key := "c"
	value := 3
	newMapPtr, err := InsertMap(arena, key, value, initialMap)
	if err != nil {
		t.Fatalf("Error inserting into map: %v", err)
	}
	if newMapPtr == nil {
		t.Fatalf("New map pointer is nil")
	}
	if (*newMapPtr)[key] != value {
		t.Errorf("Expected key %s to have value %d, got %d", key, value, (*newMapPtr)[key])
	}
}

func BenchmarkInsertMap(b *testing.B) {
	arena, err := NewMemoryArena[map[string]int](1000)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err)
	}
	initialMap := map[string]int{"a": 1, "b": 2}
	key := "c"
	value := 3
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = InsertMap(arena, key, value, initialMap)
		if err != nil {
			b.Fatalf("Error inserting into map: %v", err)
		}
	}
}
