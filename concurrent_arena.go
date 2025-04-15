package memoryArena

import (
	"sync"
	"unsafe"
)

type ConcurrentArena[T any] struct {
	*MemoryArena[T]
	mutex sync.Mutex // Exclusive access for all operations.
}

func NewConcurrentArena[T any](size int) (*ConcurrentArena[T], error) {
	arena, err := NewMemoryArena[T](size)
	if err != nil {
		return nil, err
	}
	return &ConcurrentArena[T]{
		MemoryArena: arena,
		mutex:       sync.Mutex{},
	}, nil
}

func (ca *ConcurrentArena[T]) Allocate(size int) (unsafe.Pointer, error) {
	ca.mutex.Lock()
	defer ca.mutex.Unlock()
	return ca.MemoryArena.Allocate(size)
}

func (ca *ConcurrentArena[T]) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	ca.mutex.Lock()
	defer ca.mutex.Unlock()

	value, ok := obj.(T)
	if !ok {
		return nil, ErrInvalidType
	}
	size := int(unsafe.Sizeof(value))
	ptr, err := ca.MemoryArena.Allocate(size)
	if err != nil {
		return nil, err
	}
	*(*T)(ptr) = value
	return ptr, nil
}

func (ca *ConcurrentArena[T]) ResizePreserve(newSize int) error {
	ca.mutex.Lock()
	defer ca.mutex.Unlock()
	return ca.MemoryArena.ResizePreserve(newSize)
}

func (ca *ConcurrentArena[T]) Resize(newSize int) error {
	ca.mutex.Lock()
	defer ca.mutex.Unlock()
	return ca.MemoryArena.Resize(newSize)
}
