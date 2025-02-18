package memoryArena

import (
	"fmt"
	"reflect"
	"unsafe"
)

// memory: A byte array that holds the actual memory
// size the total size of the memory arena
// offset the amount of memory currently in use
type MemoryArenaBuffer struct {
	memory []byte
	size   int
	offset int
}

// this function creates a new memory arena buffer of a specified size
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

// this function creates a new memory arena of a specified size
// it allocates a block of memory and initializes the arena's properties
func NewMemoryArena[T any](size int) (*MemoryArena[T], error) {
	if size <= 0 {
		return nil, fmt.Errorf("arena size must be greater than 0")
	}
	arena := MemoryArena[T]{
		buffer: *NewMemoryArenaBuffer(size),
	}
	return &arena, nil
}

// this function returns the remainder of the offset when divided by the alignment
func (arena *MemoryArena[T]) GetRemainder(alignment uintptr) int {
	return arena.buffer.offset % int(alignment)
}

// this function aligns the offset to the specified alignment
func (arena *MemoryArena[T]) alignOffset(alignment uintptr) {
	remainder := arena.GetRemainder(alignment)
	if remainder != 0 {
		arena.buffer.offset = (arena.buffer.offset + int(alignment-1)) &^ (int(alignment) - 1)
	}
}

// Remaining capacity of the arena
func (arena *MemoryArena[T]) nextOffset(size int) int {
	return arena.buffer.offset + size
}

// checking boundries of the arena
func (arena *MemoryArena[T]) notEnoughSpace(size int) bool {
	return arena.nextOffset(size) > arena.buffer.size
}

// this function is used to allocate memory from the arena
func (arena *MemoryArena[T]) Allocate(size int) (unsafe.Pointer, error) {
	if size <= 0 {
		return nil, fmt.Errorf("allocation size must be greater than 0")
	}
	return arena.AllocateBuffer(size)
}

// this function returns a pointer to the memory in the arena
func (arena *MemoryArena[T]) GetResult() unsafe.Pointer {
	return unsafe.Pointer(&arena.buffer.memory[arena.buffer.offset])
}

// it checks if there's enough space left in the arena
// if there is enough space, it returns a pointer to the available memory and updates the used amount
// if there is not enough space, it returns null(or some error indicator)
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

// this function frees the memory in the arena by setting all the bytes to 0
func (arena *MemoryArena[T]) Free() {
	for i := range arena.buffer.memory {
		arena.buffer.memory[i] = 0
	}
}

// this function resets the arena by setting the offset to 0
func (arena *MemoryArena[T]) Reset() {
	arena.Free()
	arena.buffer.offset = 0
}

func (arena *MemoryArena[T]) AllocateNewValue(size int, obj interface{}) (*unsafe.Pointer, error) {
	ptr, err := arena.Allocate(int(size))
	if err != nil {
		return nil, fmt.Errorf("allocation failed due to insufficient memory")
	}

	// Create a new value at the allocated memory and copy the object into it
	ptr, err = SetNewValue(&ptr, obj)
	if err != nil {
		return nil, err
	}
	return &ptr, nil
}

// AllocateObject allocates memory for the given object and returns a pointer to the allocated memory.
func (arena *MemoryArena[T]) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	size := int(reflect.TypeOf(obj).Size())
	// Allocate memory
	ptr, err := arena.AllocateNewValue(size, obj)
	if err != nil {
		return nil, err
	}
	return *ptr, nil
}

// Resize discards the old memory and reinitializes the arena with a new size.
// All previously allocated pointers become invalid after this call!
func (arena *MemoryArena[T]) Resize(newSize int) error {
	if newSize <= 0 {
		return fmt.Errorf("arena size must be greater than 0")
	}
	// Optionally clear the old memory (arena.Free()) if you want,
	// but since we're discarding it anyway, you can skip if desired.
	arena.Free()

	// Allocate a new slice with the new size
	arena.buffer.memory = make([]byte, newSize)
	arena.buffer.size = newSize

	// Reset the offset so subsequent allocations start at the beginning
	arena.buffer.offset = 0

	return nil
}

func (arena *MemoryArena[T]) ResizePreserve(newSize int) error {
	if newSize <= 0 {
		return fmt.Errorf("arena size must be greater than 0")
	}
	// Old used size
	used := arena.buffer.offset
	if used > newSize {
		// Not enough room to keep old data
		// Either return an error or shrink offset
		return fmt.Errorf("new size is smaller than current usage")
	}

	// Create new slice
	newMemory := make([]byte, newSize)
	// Copy old used bytes
	copy(newMemory, arena.buffer.memory[:used])

	// Swap in new memory
	arena.buffer.memory = newMemory
	arena.buffer.size = newSize

	// offset remains the same
	// but all old pointer addresses are still invalid
	// because they pointed to the old slice
	return nil
}
