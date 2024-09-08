package memoryArena

import (
	"sync"
	"unsafe"
)

type ConcurrentArena[T any] struct {
	*MemoryArena[T]
	mutex sync.Mutex
}

func NewConcurrentArena[T any](arena MemoryArena[T]) *ConcurrentArena[T] {
	return &ConcurrentArena[T]{
		MemoryArena: &arena,
	}
}

func (arena *ConcurrentArena[T]) Allocate(size int) (unsafe.Pointer, error) {
	arena.mutex.Lock()
	ptr, err := arena.MemoryArena.Allocate(size)
	if err != nil {
		return nil, err
	}
	defer arena.mutex.Unlock()
	return ptr, nil
}

func (arena *ConcurrentArena[T]) Reset() {
	arena.mutex.Lock()
	arena.MemoryArena.Reset()
	defer arena.mutex.Unlock()
}

func (arena *ConcurrentArena[T]) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	arena.mutex.Lock()
	val, err := arena.MemoryArena.AllocateObject(obj)
	if err != nil {
		return nil, err
	}
	defer arena.mutex.Unlock()
	return val, nil
}
