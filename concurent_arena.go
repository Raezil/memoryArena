package memoryArena

import (
	"sync"
	"unsafe"
)

type ConcurrentArena[T any] struct {
	mutex sync.Mutex
	arena MemoryArena[T]
}

func NewConcurrentArena[T any](arena MemoryArena[T]) *ConcurrentArena[T] {
	return &ConcurrentArena[T]{
		arena: arena,
	}
}

func (concurrentArena *ConcurrentArena[T]) Allocate(size int) unsafe.Pointer {
	concurrentArena.mutex.Lock()
	ptr := concurrentArena.arena.Allocate(size)
	concurrentArena.mutex.Unlock()
	return ptr
}

func (concurrentArena *ConcurrentArena[T]) Reset() {
	concurrentArena.mutex.Lock()
	concurrentArena.arena.Reset()
	concurrentArena.mutex.Unlock()
}

func (concurrentArena *ConcurrentArena[T]) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	concurrentArena.mutex.Lock()
	val, err := concurrentArena.arena.AllocateObject(obj)
	if err != nil {
		return nil, err
	}
	concurrentArena.mutex.Unlock()
	return val, nil
}
