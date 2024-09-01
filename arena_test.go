package memoryArena

import "testing"

func TestMemoryArena(t *testing.T) {
	arena := NewArena(1024)

	_, err := arena.AllocateObject(int(42))
	if err != nil {
		t.Fatalf("Allocation failed: %v", err)
	}

	arena.Reset()

	_, err = arena.AllocateObject(int(7))
	if err != nil {
		t.Fatalf("Allocation failed: %v", err)
	}
}
