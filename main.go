package main

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
	used   int
}

// this function creates a new memory arena of a specified size
// it allocates a block of memory and initializes the arena's properties
func NewArena(size int) *MemoryArena {
	arena := MemoryArena{
		memory: make([]byte, size),
		size:   size,
		used:   0,
	}
	return &arena
}

// this function is used to allocate memory from the arena
// it checks if there's enough space left in the arena
// if there is enough space, it returns a pointer to the available memory and updates the used amount
// if there is not enough space, it returns null(or some error indicator)
func (arena *MemoryArena) Allocate(size int) unsafe.Pointer {
	if arena.used+size > arena.size {
		return nil
	}
	result := unsafe.Pointer(&arena.memory[arena.used])
	arena.used += size
	return result
}

func (arena *MemoryArena) Reset() {
	arena.used = 0
}

func (arena *MemoryArena) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	size := reflect.TypeOf(obj).Size()
	ptr := arena.Allocate(int(size))
	if ptr == nil {
		return nil, fmt.Errorf("cannot procees further")
	}
	reflect.NewAt(
		reflect.TypeOf(obj),
		ptr,
	)
	return ptr, nil
}
