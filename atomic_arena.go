package memoryArena

import (
	"sync/atomic"
	"unsafe"
)

// -------------------------------------------------------------
// AtomicMemoryArena provides lock-free, thread-safe bump allocation
// using the existing MemoryArenaBuffer.
type AtomicArena[T any] struct {
	buffer    *MemoryArenaBuffer
	alignment uintptr
	offset    atomic.Uintptr
}

// NewAtomicMemoryArena creates a new AtomicMemoryArena with the specified capacity.
// Returns ErrInvalidSize if size <= 0.
func NewAtomicArena[T any](size int) (*AtomicArena[T], error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}
	var zero T
	alignment := unsafe.Alignof(zero)
	buf := NewMemoryArenaBuffer(size, alignment)
	arena := &AtomicArena[T]{
		buffer:    buf,
		alignment: alignment,
	}
	// initialize offset to the base alignment offset
	arena.offset.Store(uintptr(buf.offset))
	return arena, nil
}

// Allocate atomically reserves size bytes and returns a pointer, or an error.
func (arena *AtomicArena[T]) Allocate(size int) (unsafe.Pointer, error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}
	for {
		old := arena.offset.Load()
		// align the old offset
		rem := old % arena.alignment
		var aligned uintptr
		if rem != 0 {
			aligned = old + (arena.alignment - rem)
		} else {
			aligned = old
		}
		newOff := aligned + uintptr(size)
		if int(newOff) > arena.buffer.size {
			return nil, ErrArenaFull
		}
		// try to reserve
		if arena.offset.CompareAndSwap(old, newOff) {
			return unsafe.Pointer(&arena.buffer.memory[aligned]), nil
		}
		// retry on contention
	}
}

// AllocateNewValue atomically allocates and copies obj into the arena.
func (arena *AtomicArena[T]) AllocateNewValue(obj T) (unsafe.Pointer, error) {
	sz := int(unsafe.Sizeof(obj))
	ptr, err := arena.Allocate(sz)
	if err != nil {
		return nil, err
	}
	*(*T)(ptr) = obj
	return ptr, nil
}

// Reset clears the buffer and resets the offset to the base alignment.
func (arena *AtomicArena[T]) Reset() {
	// zero out memory
	for i := range arena.buffer.memory {
		arena.buffer.memory[i] = 0
	}
	// reset atomic offset
	arena.offset.Store(uintptr(arena.buffer.offset))
}
