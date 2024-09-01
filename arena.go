package memoryArena

import (
	"fmt"
	"reflect"
	"unsafe"
)

// memory: A byte array that holds the actual memory
// size the total size of the memory arena
// used the amount of memory currently in use
type MemoryArena struct {
	memory []byte
	size   int
	offset int
}

// this function creates a new memory arena of a specified size
// it allocates a block of memory and initializes the arena's properties
func NewArena(size int) *MemoryArena {
	arena := MemoryArena{
		memory: make([]byte, size),
		size:   size,
		offset: 0,
	}
	return &arena
}

// this function is used to allocate memory from the arena
// it checks if there's enough space left in the arena
// if there is enough space, it returns a pointer to the available memory and updates the used amount
// if there is not enough space, it returns null(or some error indicator)
func (arena *MemoryArena) Allocate(size int) unsafe.Pointer {
	if arena.offset+size > arena.size {
		return nil
	}
	result := unsafe.Pointer(&arena.memory[arena.offset])
	arena.offset += size
	return result
}

func (arena *MemoryArena) Reset() {
	arena.offset = 0
	for i := range arena.memory {
		arena.memory[i] = 0
	}
}

// AllocateObject allocates memory for the given object and returns a pointer to the allocated memory.
func (arena *MemoryArena) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	size := reflect.TypeOf(obj).Size()

	// Allocate memory
	ptr := arena.Allocate(int(size))
	if ptr == nil {
		return nil, fmt.Errorf("allocation failed due to insufficient memory")
	}

	// Create a new value at the allocated memory and copy the object into it
	newValue := reflect.NewAt(
		reflect.TypeOf(obj),
		ptr,
	).Elem()
	newValue.Set(reflect.ValueOf(obj))
	return ptr, nil
}

// NewObject ollocate memory through AllocateObject, returns pointer to T or error handle.
func NewObject[T any](arena *MemoryArena, obj T) (*T, error) {
	ptr, err := arena.AllocateObject(obj)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}
