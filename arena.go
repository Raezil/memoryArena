package memoryArena

import "unsafe"

type Arena[T any] interface {
	// Allocate reserves sz bytes (aligned for T) and returns a pointer to the start.
	Allocate(sz int) (unsafe.Pointer, error)
	// NewObject allocates space for one T, copies obj into it, and returns *T.
	NewObject(obj T) (*T, error)
	// Reset clears all allocations and resets the arena to empty state.
	Reset()
	// AppendSlice appends elems to an existing slice, growing in-arena if needed.
	AppendSlice(slice []T, elems ...T) ([]T, error)
	Offset() int
	Base() unsafe.Pointer
}
