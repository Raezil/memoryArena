package memoryArena

import (
	"sync"
	"unsafe"
)

type ConcurrentArena[T any] struct {
	mu    sync.Mutex
	arena Arena[T]
}

func NewConcurrentArena[T any](size int) (Arena[T], error) {
	a, err := NewMemoryArena[T](size)
	if err != nil {
		return nil, err
	}
	return &ConcurrentArena[T]{arena: a}, nil
}

func (c *ConcurrentArena[T]) Allocate(sz int) (unsafe.Pointer, error) {
	c.mu.Lock()
	p, err := c.arena.Allocate(sz)
	c.mu.Unlock()
	return p, err
}

func (c *ConcurrentArena[T]) NewObject(obj T) (*T, error) {
	c.mu.Lock()
	ptr, err := c.arena.NewObject(obj)
	c.mu.Unlock()
	return ptr, err
}

func (c *ConcurrentArena[T]) AppendSlice(slice []T, elems ...T) ([]T, error) {
	c.mu.Lock()
	out, err := c.arena.AppendSlice(slice, elems...)
	c.mu.Unlock()
	return out, err
}

func (c *ConcurrentArena[T]) Reset() {
	c.mu.Lock()
	c.arena.Reset()
	c.mu.Unlock()
}

func (c *ConcurrentArena[T]) Offset() int {
	return c.arena.Offset()
}

func (ca *ConcurrentArena[T]) Base() unsafe.Pointer {
	return ca.arena.Base()
}
