package memoryArena

import (
	"fmt"
	"sync"
	"unsafe"
)

type ConcurrentArena[T any] struct {
	*MemoryArena[T]
	mutex sync.RWMutex
}

func NewConcurrentArena[T any](size int) (*ConcurrentArena[T], error) {
	arena, err := NewMemoryArena[T](size)
	if err != nil {
		return nil, err
	}
	return &ConcurrentArena[T]{
		MemoryArena: arena,
		mutex:       sync.RWMutex{},
	}, nil
}

func (arena *ConcurrentArena[T]) Allocate(size int) (unsafe.Pointer, error) {
	return arena.MemoryArena.Allocate(size)
}

func (arena *ConcurrentArena[T]) Reset() {
	arena.mutex.Lock()
	defer arena.mutex.Unlock()
	arena.MemoryArena.Reset()
}

func (arena *ConcurrentArena[T]) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	arena.mutex.Lock()
	defer arena.mutex.Unlock()

	// Ensure the object is of the expected type.
	value, ok := obj.(T)
	if !ok {
		return nil, fmt.Errorf("invalid type: expected %T", *new(T))
	}
	size := int(unsafe.Sizeof(value))
	ptr, err := arena.Allocate(size)
	if err != nil {
		return nil, fmt.Errorf("allocation failed due to insufficient memory")
	}
	// Directly copy the value into the allocated memory.
	*(*T)(ptr) = value
	return ptr, nil
}

func (arena *ConcurrentArena[T]) ResizePreserve(newSize int) error {
	arena.mutex.Lock()
	defer arena.mutex.Unlock()
	return arena.MemoryArena.ResizePreserve(newSize)
}

func (arena *ConcurrentArena[T]) Resize(newSize int) error {
	arena.mutex.Lock()
	defer arena.mutex.Unlock()
	return arena.MemoryArena.Resize(newSize)
}
