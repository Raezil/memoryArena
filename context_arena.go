package memoryArena

import (
	"context"
	"unsafe"
)

// ContextArena wraps MemoryArena and resets its state when the provided context is canceled.
// It proxies Allocate, NewObject, and AppendSlice calls, returning context errors if the context is done.

type ContextArena[T any] struct {
	arena *MemoryArena[T]
	ctx   context.Context
}

// NewContextArena creates a new MemoryArena of the given size and ties its lifecycle to ctx.
// When ctx is canceled, the underlying arena is automatically Reset.
func NewContextArena[T any](ctx context.Context, size int) (*ContextArena[T], error) {
	ma, err := NewMemoryArena[T](size)
	if err != nil {
		return nil, err
	}
	ca := &ContextArena[T]{arena: ma, ctx: ctx}
	go func() {
		<-ctx.Done()
		ma.Reset()
	}()
	return ca, nil
}

// Allocate reserves sz bytes from the arena, unless the context is done.
func (c *ContextArena[T]) Allocate(sz int) (unsafe.Pointer, error) {
	select {
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	default:
		return c.arena.Allocate(sz)
	}
}

// NewObject allocates space for obj and returns its pointer, or a context error if ctx is done.
func (c *ContextArena[T]) NewObject(obj T) (*T, error) {
	select {
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	default:
		return c.arena.NewObject(obj)
	}
}

// AppendSlice extends slice with elems in the arena, or returns a context error if ctx is done.
func (c *ContextArena[T]) AppendSlice(slice []T, elems ...T) ([]T, error) {
	select {
	case <-c.ctx.Done():
		return slice, c.ctx.Err()
	default:
		return c.arena.AppendSlice(slice, elems...)
	}
}

// Reset immediately resets the arena, regardless of context state.
func (c *ContextArena[T]) Reset() {
	c.arena.Reset()
}
