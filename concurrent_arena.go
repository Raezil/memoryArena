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
	arena.mutex.Lock()
	defer arena.mutex.Unlock()
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
	size := reflect.TypeOf(obj).Size()
	defer arena.mutex.Unlock()

	// Allocate memory
	ptr, err := arena.Allocate(int(size))
	if err != nil {
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
