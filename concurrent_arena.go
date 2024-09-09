package memoryArena

import (
	"fmt"
	"sync"
	"unsafe"
)

type ConcurrentArena[T any] struct {
	// embedding MemoryArena type in ConcurrentArena
	*MemoryArena[T]
	mutex sync.Mutex
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
	arena.MemoryArena.Reset()
	arena.mutex.Unlock()
}

// Object is being allocated in the Concurrent Arena.
func (arena *ConcurrentArena[T]) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	arena.mutex.Lock()
	ptr, err := arena.AllocateObject(obj)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	arena.mutex.Unlock()

	return ptr, nil
}
