package memoryArena

import (
	"sync"
	"unsafe"
)

// ConcurrentArena wraps MemoryArena with a mutex for thread-safe operations.
type ConcurrentArena[T any] struct {
	*MemoryArena[T]
	mutex sync.RWMutex
}

// NewConcurrentArena creates a new ConcurrentArena.
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

// Allocate performs a thread-safe allocation.
func (arena *ConcurrentArena[T]) Allocate(size int) (unsafe.Pointer, error) {
	arena.mutex.Lock()
	defer arena.mutex.Unlock()
	return arena.MemoryArena.Allocate(size)
}

// AllocateNewValue performs a thread-safe allocation of a new value.
func (arena *ConcurrentArena[T]) AllocateNewValue(obj T) (*T, error) {
	arena.mutex.Lock()
	defer arena.mutex.Unlock()
	return arena.MemoryArena.AllocateNewValue(obj)
}

// Reset safely resets the arena.
func (arena *ConcurrentArena[T]) Reset() {
	arena.mutex.Lock()
	defer arena.mutex.Unlock()
	arena.MemoryArena.Reset()
}

// Free safely frees the arena.
func (arena *ConcurrentArena[T]) Free() {
	arena.mutex.Lock()
	defer arena.mutex.Unlock()
	arena.MemoryArena.Free()
}

// ResizePreserve safely resizes the arena while preserving data.
func (arena *ConcurrentArena[T]) ResizePreserve(newSize int) error {
	arena.mutex.Lock()
	defer arena.mutex.Unlock()
	return arena.MemoryArena.ResizePreserve(newSize)
}

// Resize safely resizes the arena.
func (arena *ConcurrentArena[T]) Resize(newSize int) error {
	arena.mutex.Lock()
	defer arena.mutex.Unlock()
	return arena.MemoryArena.Resize(newSize)
}

// GetResult safely retrieves the result pointer.
func (arena *ConcurrentArena[T]) GetResult() unsafe.Pointer {
	arena.mutex.RLock()
	defer arena.mutex.RUnlock()
	return arena.MemoryArena.GetResult()
}
