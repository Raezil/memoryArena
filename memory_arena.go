package memoryArena

import (
	"fmt"
	"reflect"
	"unsafe"
)

// memory: A byte array that holds the actual memory
// size the total size of the memory arena
// offset the amount of memory currently in use
type MemoryArena[T any] struct {
	memory []byte
	size   int
	offset int
}

// this function creates a new memory arena of a specified size
// it allocates a block of memory and initializes the arena's properties
func NewMemoryArena[T any](size int) (*MemoryArena[T], error) {
	if size <= 0 {
		return nil, fmt.Errorf("Cannot initialize, size below 0")
	}
	arena := MemoryArena[T]{
		memory: make([]byte, size),
		size:   size,
		offset: 0,
	}
	return &arena, nil
}

func (arena *MemoryArena[T]) alignOffset(alignment uintptr) {
	if (arena.offset % int(alignment)) != 0 {
		arena.offset = (arena.offset + int(alignment-1)) &^ (int(alignment) - 1)
	}
}

// this function is used to allocate memory from the arena
// it checks if there's enough space left in the arena
// if there is enough space, it returns a pointer to the available memory and updates the used amount
// if there is not enough space, it returns null(or some error indicator)
func (arena *MemoryArena[T]) Allocate(size int) (unsafe.Pointer, error) {
	if size <= 0 {
		return nil, fmt.Errorf("allocation size must be greater than 0")
	}

	alignment := unsafe.Alignof(new(T))
	arena.alignOffset(alignment)

	if arena.offset+size > arena.size {
		return nil, fmt.Errorf("not enough space left in the arena")
	}

	result := unsafe.Pointer(&arena.memory[arena.offset])
	arena.offset += size
	return result, nil
}

func (arena *MemoryArena[T]) Free() {
	for i := range arena.memory {
		arena.memory[i] = 0
	}
}

func (arena *MemoryArena[T]) Reset() {
	arena.offset = 0
	arena.Free()
}

// AllocateObject allocates memory for the given object and returns a pointer to the allocated memory.
func (arena *MemoryArena[T]) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	size := reflect.TypeOf(obj).Size()
	// Allocate memory
	ptr, err := arena.Allocate(int(size))
	if err != nil {
		return nil, fmt.Errorf("allocation failed due to insufficient memory")
	}

	// Create a new value at the allocated memory and copy the object into it
	ptr = SetNewValue(&ptr, obj)
	return ptr, nil
}
