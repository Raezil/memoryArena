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
	arena := NewMemoryArenaBuffer(100)
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
