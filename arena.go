package memoryArena

import "unsafe"

// interface for MemoryArena and ConcurrentArena behaviours
type Arena interface {
	Reset()
	AllocateObject(obj interface{}) (unsafe.Pointer, error)
}

// Object is being allocated in the Arena.
func AllocateObject[T any](arena Arena, obj T) (unsafe.Pointer, error) {
	return arena.AllocateObject(obj)
}

func Reset(arena Arena) {
	arena.Reset()
}

// NewObject allocate memory through AllocateObject, returns pointer to T or error handle.
func NewObject[T any](arena Arena, obj T) (*T, error) {
	ptr, err := AllocateObject(arena, obj)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}
