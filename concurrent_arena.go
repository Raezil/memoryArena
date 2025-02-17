package memoryArena

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

type ConcurrentArena[T any] struct {
	// embedding MemoryArena type in ConcurrentArena
	*MemoryArena[T]
	mutex sync.RWMutex
}

// Constructor of ConcurrentArena
func NewConcurrentArena[T any](arena MemoryArena[T]) *ConcurrentArena[T] {
	return &ConcurrentArena[T]{
		MemoryArena: &arena,
	}
}

// Allocating object in conccurrent arena
func (arena *ConcurrentArena[T]) Allocate(size int) (unsafe.Pointer, error) {
	ptr, err := arena.MemoryArena.Allocate(size)
	if err != nil {
		return nil, err
	}
	return ptr, nil
}

// Resetting concurrent arena
func (arena *ConcurrentArena[T]) Reset() {
	arena.mutex.Lock()
	defer arena.mutex.Unlock()
	arena.MemoryArena.Reset()
}

// Object is being allocated in the Concurrent Arena.
func (arena *ConcurrentArena[T]) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	arena.mutex.Lock()
	defer arena.mutex.Unlock()
	// Get the size of the object
	size := reflect.TypeOf(obj).Size()
	// Allocate memory
	ptr, err := arena.Allocate(int(size))
	if err != nil {
		return nil, fmt.Errorf("allocation failed due to insufficient memory")
	}

	// Create a new value at the allocated memory and copy the object into it
	ptr, err = SetNewValue(&ptr, obj)
	if err != nil {
		return nil, err
	}
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
