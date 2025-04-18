package memoryArena

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

// arenaState holds internal state for AtomicArena, including buffer and current offset.
type arenaState[T any] struct {
	buffer *MemoryArenaBuffer
	offset uintptr
}

// AtomicArena provides lock-free, thread-safe bump allocation using MemoryArenaBuffer.
type AtomicArena[T any] struct {
	state     atomic.Pointer[arenaState[T]]
	alignment uintptr
}

// NewAtomicArena creates a new AtomicArena with the specified capacity.
// Returns ErrInvalidSize if size <= 0.
func NewAtomicArena[T any](size int) (*AtomicArena[T], error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}
	var zero T
	alignment := unsafe.Alignof(zero)
	buf := NewMemoryArenaBuffer(size, alignment)
	arena := &AtomicArena[T]{alignment: alignment}
	initial := &arenaState[T]{
		buffer: buf,
		offset: uintptr(buf.offset),
	}
	arena.state.Store(initial)
	return arena, nil
}

// Allocate atomically reserves size bytes and returns a pointer, or an error.
func (arena *AtomicArena[T]) Allocate(size int) (unsafe.Pointer, error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}
	for {
		old := arena.state.Load()
		buf := old.buffer
		off := old.offset
		// align the offset
		rem := off % arena.alignment
		var aligned uintptr
		if rem != 0 {
			aligned = off + (arena.alignment - rem)
		} else {
			aligned = off
		}
		newOff := aligned + uintptr(size)
		// check capacity
		if int(newOff) > buf.size {
			return nil, ErrArenaFull
		}
		// attempt to update state
		newState := &arenaState[T]{buffer: buf, offset: newOff}
		if arena.state.CompareAndSwap(old, newState) {
			// Audited: conversion of &buf.memory[aligned] to unsafe.Pointer is safe
			// because `aligned` has been validated against buf.size above.
			// #nosec G103
			return unsafe.Pointer(&buf.memory[aligned]), nil
		}
		// retry on contention
	}
}

// AllocateNewValue allocates and copies obj into the arena.
func (arena *AtomicArena[T]) AllocateNewValue(obj T) (unsafe.Pointer, error) {
	sz := int(unsafe.Sizeof(obj))
	ptr, err := arena.Allocate(sz)
	if err != nil {
		return nil, err
	}
	*(*T)(ptr) = obj
	return ptr, nil
}

// AllocateObject allocates and initializes the provided obj (of expected type T) in the arena.
func (arena *AtomicArena[T]) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	typedObj, ok := obj.(T)
	if !ok {
		return nil, fmt.Errorf("AllocateObject: expected type %T, got %T", *new(T), obj)
	}
	return arena.AllocateNewValue(typedObj)
}

// Reset clears the arena by replacing its buffer, safe under concurrent allocations.
func (arena *AtomicArena[T]) Reset() {
	old := arena.state.Load()
	bufSize := old.buffer.size
	newBuf := NewMemoryArenaBuffer(bufSize, arena.alignment)
	newState := &arenaState[T]{buffer: newBuf, offset: uintptr(newBuf.offset)}
	arena.state.Store(newState)
}

// Resize resets the arena to newSize, discarding all data. Safe under concurrent allocations.
func (arena *AtomicArena[T]) Resize(newSize int) error {
	if newSize <= 0 {
		return ErrInvalidSize
	}
	newBuf := NewMemoryArenaBuffer(newSize, arena.alignment)
	newState := &arenaState[T]{buffer: newBuf, offset: uintptr(newBuf.offset)}
	arena.state.Store(newState)
	return nil
}

// ResizePreserve resizes the arena to newSize, preserving existing data.
// Note: not safe under concurrent allocations.
func (arena *AtomicArena[T]) ResizePreserve(newSize int) error {
	if newSize <= 0 {
		return ErrInvalidSize
	}
	for {
		old := arena.state.Load()
		buf := old.buffer
		// compute how many bytes are in use, beyond the initial alignment
		base := uintptr(buf.offset)
		used := int(old.offset - base)
		if used > newSize {
			return ErrNewSizeTooSmall
		}
		// allocate new buffer and copy the used portion
		newBuf := NewMemoryArenaBuffer(newSize, arena.alignment)
		copy(newBuf.memory[newBuf.offset:], buf.memory[buf.offset:buf.offset+used])
		newState := &arenaState[T]{buffer: newBuf, offset: uintptr(newBuf.offset) + uintptr(used)}
		// attempt CAS; if it succeeds, we've atomically swapped to the new preserved state
		if arena.state.CompareAndSwap(old, newState) {
			return nil
		}
		// retry: another goroutine allocated meanwhile, so reload and preserve a larger snapshot
	}
}
