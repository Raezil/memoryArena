package memoryArena

import (
	"fmt"
	"unsafe"
)

type MemoryArenaBuffer struct {
	memory []byte
	size   int
	offset int
}

func NewMemoryArenaBuffer(size int) *MemoryArenaBuffer {
	return &MemoryArenaBuffer{
		memory: make([]byte, size),
		size:   size,
		offset: 0,
	}
}

type MemoryArena[T any] struct {
	buffer MemoryArenaBuffer
}

func NewMemoryArena[T any](size int) (*MemoryArena[T], error) {
	if size <= 0 {
		return nil, fmt.Errorf("arena size must be greater than 0")
	}
	arena := MemoryArena[T]{
		buffer: *NewMemoryArenaBuffer(size),
	}
	return &arena, nil
}

func (arena *MemoryArena[T]) GetRemainder(alignment uintptr) int {
	return arena.buffer.offset % int(alignment)
}

func (arena *MemoryArena[T]) alignOffset(alignment uintptr) {
	if remainder := arena.GetRemainder(alignment); remainder != 0 {
		arena.buffer.offset += int(alignment) - remainder
	}
}

func (arena *MemoryArena[T]) nextOffset(size int) int {
	return arena.buffer.offset + size
}

func (arena *MemoryArena[T]) notEnoughSpace(size int) bool {
	return arena.nextOffset(size) > arena.buffer.size
}

func (arena *MemoryArena[T]) Allocate(size int) (unsafe.Pointer, error) {
	if size <= 0 {
		return nil, fmt.Errorf("allocation size must be greater than 0")
	}
	return arena.AllocateBuffer(size)
}

func (arena *MemoryArena[T]) GetResult() unsafe.Pointer {
	return unsafe.Pointer(&arena.buffer.memory[arena.buffer.offset])
}

func (arena *MemoryArena[T]) AllocateBuffer(size int) (unsafe.Pointer, error) {
	alignment := unsafe.Alignof(new(T))
	arena.alignOffset(alignment)
	if arena.notEnoughSpace(size) {
		return nil, fmt.Errorf("not enough space left in the arena")
	}
	result := arena.GetResult()
	arena.buffer.offset += size
	return result, nil
}

func (arena *MemoryArena[T]) Free() {
	for i := range arena.buffer.memory {
		arena.buffer.memory[i] = 0
	}
}

func (arena *MemoryArena[T]) Reset() {
	arena.Free()
	arena.buffer.offset = 0
}

// AllocateNewValue allocates space and copies the provided object into that space.
func (arena *MemoryArena[T]) AllocateNewValue(size int, obj T) (*unsafe.Pointer, error) {
	ptr, err := arena.Allocate(size)
	if err != nil {
		return nil, fmt.Errorf("allocation failed due to insufficient memory")
	}
	*(*T)(ptr) = obj
	return &ptr, nil
}

// AllocateObject allocates memory for an object and copies its value into the arena.
func (arena *MemoryArena[T]) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	value, ok := obj.(T)
	if !ok {
		return nil, fmt.Errorf("invalid type: expected %T", *new(T))
	}
	size := int(unsafe.Sizeof(value))
	ptr, err := arena.AllocateNewValue(size, value)
	if err != nil {
		return nil, err
	}
	return *ptr, nil
}

func (arena *MemoryArena[T]) Resize(newSize int) error {
	if newSize <= 0 {
		return fmt.Errorf("arena size must be greater than 0")
	}
	arena.Free()
	arena.buffer.memory = make([]byte, newSize)
	arena.buffer.size = newSize
	arena.buffer.offset = 0
	return nil
}

func (arena *MemoryArena[T]) ResizePreserve(newSize int) error {
	if newSize <= 0 {
		return fmt.Errorf("arena size must be greater than 0")
	}
	used := arena.buffer.offset
	if used > newSize {
		return fmt.Errorf("new size is smaller than current usage")
	}
	arena.SetNewMemory(newSize, used)
	return nil
}

func (arena *MemoryArena[T]) SetNewMemory(newSize int, used int) {
	newMemory := make([]byte, newSize)
	copy(newMemory, arena.buffer.memory[:used])
	arena.buffer.memory = newMemory
	arena.buffer.size = newSize
}
