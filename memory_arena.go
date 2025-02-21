package memoryArena

import (
	"fmt"
	"unsafe"
)

// MemoryArenaBuffer holds the underlying memory.
type MemoryArenaBuffer struct {
	memory []byte
	size   int
	offset int
}

// NewMemoryArenaBuffer creates a new MemoryArenaBuffer with the specified size.
func NewMemoryArenaBuffer(size int) *MemoryArenaBuffer {
	return &MemoryArenaBuffer{
		memory: make([]byte, size),
		size:   size,
		offset: 0,
	}
}

// MemoryArena provides low-level allocation from a fixed-size buffer.
type MemoryArena[T any] struct {
	buffer *MemoryArenaBuffer
}

// NewMemoryArena creates a new MemoryArena.
func NewMemoryArena[T any](size int) (*MemoryArena[T], error) {
	if size <= 0 {
		return nil, fmt.Errorf("arena size must be greater than 0")
	}
	return &MemoryArena[T]{
		buffer: NewMemoryArenaBuffer(size),
	}, nil
}

// alignOffset adjusts the buffer offset to satisfy the alignment requirement.
func (arena *MemoryArena[T]) alignOffset(alignment uintptr) {
	remainder := arena.buffer.offset % int(alignment)
	if remainder != 0 {
		arena.buffer.offset += int(alignment) - remainder
	}
}

// hasEnoughSpace checks if there is room for size bytes.
func (arena *MemoryArena[T]) hasEnoughSpace(size int) bool {
	return arena.buffer.offset+size <= arena.buffer.size
}

// Allocate reserves a block of memory of the given size.
func (arena *MemoryArena[T]) Allocate(size int) (unsafe.Pointer, error) {
	if size <= 0 {
		return nil, fmt.Errorf("allocation size must be greater than 0")
	}
	alignment := unsafe.Alignof(*new(T))
	arena.alignOffset(alignment)
	if !arena.hasEnoughSpace(size) {
		return nil, fmt.Errorf("not enough space left in the arena")
	}
	ptr := unsafe.Pointer(&arena.buffer.memory[arena.buffer.offset])
	arena.buffer.offset += size
	return ptr, nil
}

// AllocateNewValue allocates memory for an object of type T, copies the value, and returns a pointer to it.
func (arena *MemoryArena[T]) AllocateNewValue(obj T) (*T, error) {
	size := int(unsafe.Sizeof(obj))
	ptr, err := arena.Allocate(size)
	if err != nil {
		return nil, err
	}
	newObj := (*T)(ptr)
	*newObj = obj
	return newObj, nil
}

// Reset clears the arena and resets the offset.
func (arena *MemoryArena[T]) Reset() {
	for i := range arena.buffer.memory {
		arena.buffer.memory[i] = 0
	}
	arena.buffer.offset = 0
}

// Free clears the memory (an alias for Reset in this simple implementation).
func (arena *MemoryArena[T]) Free() {
	arena.Reset()
}

// Resize discards the old memory and allocates a new block.
func (arena *MemoryArena[T]) Resize(newSize int) error {
	if newSize <= 0 {
		return fmt.Errorf("arena size must be greater than 0")
	}
	arena.buffer.memory = make([]byte, newSize)
	arena.buffer.size = newSize
	arena.buffer.offset = 0
	return nil
}

// ResizePreserve resizes the arena while preserving existing data.
func (arena *MemoryArena[T]) ResizePreserve(newSize int) error {
	if newSize <= 0 {
		return fmt.Errorf("arena size must be greater than 0")
	}
	used := arena.buffer.offset
	if used > newSize {
		return fmt.Errorf("new size is smaller than current usage")
	}
	newMemory := make([]byte, newSize)
	copy(newMemory, arena.buffer.memory[:used])
	arena.buffer.memory = newMemory
	arena.buffer.size = newSize
	return nil
}

// GetResult returns a pointer to the next free memory location.
func (arena *MemoryArena[T]) GetResult() unsafe.Pointer {
	return unsafe.Pointer(&arena.buffer.memory[arena.buffer.offset])
}
