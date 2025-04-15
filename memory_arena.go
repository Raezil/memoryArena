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

func NewMemoryArenaBuffer(size int, alignment uintptr) *MemoryArenaBuffer {
	// Allocate extra space to ensure you can align the base.
	buf := make([]byte, size+int(alignment))

	// Compute an aligned offset from the beginning of buf.
	base := uintptr(unsafe.Pointer(&buf[0]))
	offset := 0
	if rem := base % alignment; rem != 0 {
		offset = int(alignment - rem)
	}

	// The effective usable size is reduced by the initial offset.
	return &MemoryArenaBuffer{
		memory: buf,
		size:   size, // we want "size" bytes of usable space
		offset: offset,
	}
}

type MemoryArena[T any] struct {
	buffer MemoryArenaBuffer
}

func NewMemoryArena[T any](size int) (*MemoryArena[T], error) {
	if size <= 0 {
		return nil, fmt.Errorf("arena size must be greater than 0")
	}
	alignment := unsafe.Alignof(*new(T))
	buffer := NewMemoryArenaBuffer(size, alignment)
	arena := MemoryArena[T]{
		buffer: *buffer,
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
	// Use the alignment of T itself.
	alignment := unsafe.Alignof(*new(T))
	arena.alignOffset(alignment)
	if arena.notEnoughSpace(size) {
		return nil, ErrArenaFull
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
		return nil, err
	}
	*(*T)(ptr) = obj
	return &ptr, nil
}

// AllocateObject allocates memory for an object and copies its value into the arena.
func (arena *MemoryArena[T]) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	value, ok := obj.(T)
	if !ok {
		return nil, ErrInvalidType
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
		return ErrInvalidSize
	}
	arena.Free()
	arena.buffer.memory = make([]byte, newSize)
	arena.buffer.size = newSize
	arena.buffer.offset = 0
	return nil
}

func (arena *MemoryArena[T]) ResizePreserve(newSize int) error {
	if newSize <= 0 {
		return ErrInvalidSize
	}
	used := arena.buffer.offset
	if used > newSize {
		return ErrNewSizeTooSmall
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
