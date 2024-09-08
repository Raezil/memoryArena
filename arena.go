package memoryArena

import "unsafe"

// interface for MemoryArena and ConcurrentArena behaviours
type Arena interface {
	Allocate(size int) (unsafe.Pointer, error)
	Reset()
	AllocateObject(obj interface{}) (unsafe.Pointer, error)
}

// Arena allocates memory space to later use it
func Allocate(arena Arena, size int) (unsafe.Pointer, error) {
	return arena.Allocate(size)
}

// Object is being allocated in the Arena.
func AllocateObject(arena Arena, obj interface{}) (unsafe.Pointer, error) {
	return arena.AllocateObject(obj)
}

func Reset(arena Arena) {
	arena.Reset()
}

// NewObject ollocate memory through AllocateObject, returns pointer to T or error handle.
func NewObject[T any](arena Arena, obj T) (*T, error) {
	ptr, err := arena.AllocateObject(obj)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}
