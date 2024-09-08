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

func (arena *ConcurrentArena[T]) Allocate(size int) unsafe.Pointer {
	arena.mutex.Lock()
	ptr := arena.arena.Allocate(size)
	arena.mutex.Unlock()
	return ptr
}

func (arena *ConcurrentArena[T]) Reset() {
	arena.mutex.Lock()
	arena.arena.Reset()
	arena.mutex.Unlock()
}

func (arena *ConcurrentArena[T]) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	return arena.arena.AllocateObject(obj)
}
