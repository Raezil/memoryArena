package memoryArena

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

// -------------------------------------------------------------
// AtomicArena provides lock-free, thread-safe bump allocation
// using the existing MemoryArenaBuffer.
// -------------------------------------------------------------
type AtomicArena[T any] struct {
	buffer    *MemoryArenaBuffer
	alignment uintptr
	offset    atomic.Uintptr
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

// AllocateObject allocates and initializes the provided obj (of expected type T) in the arena.
// Returns a raw unsafe.Pointer to the stored value.
func (arena *AtomicArena[T]) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	// ensure the object matches the arena's type
	typedObj, ok := obj.(T)
	if !ok {
		return nil, fmt.Errorf("AllocateObject: expected type %T, got %T", *new(T), obj)
	}
	// delegate to AllocateNewValue for allocation and copy
	return arena.AllocateNewValue(typedObj)
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

// Resize resets the arena to newSize, discarding all data.
// Returns ErrInvalidSize if newSize <= 0.
func (arena *AtomicArena[T]) Resize(newSize int) error {
	if newSize <= 0 {
		return ErrInvalidSize
	}
	// create a new buffer with the same alignment
	newBuf := NewMemoryArenaBuffer(newSize, arena.alignment)
	arena.buffer = newBuf
	// reset offset to base offset
	arena.offset.Store(uintptr(newBuf.offset))
	return nil
}

// ResizePreserve resizes the arena to newSize, preserving existing data.
// Returns ErrInvalidSize if newSize <= 0 or ErrNewSizeTooSmall if used > newSize.
func (arena *AtomicArena[T]) ResizePreserve(newSize int) error {
	if newSize <= 0 {
		return ErrInvalidSize
	}
	// compute how many bytes have been used so far
	oldOff := arena.offset.Load()
	base := uintptr(arena.buffer.offset)
	used := int(oldOff - base)
	if used > newSize {
		return ErrNewSizeTooSmall
	}
	// create new buffer
	newBuf := NewMemoryArenaBuffer(newSize, arena.alignment)
	// copy old data into the new buffer
	copy(newBuf.memory[newBuf.offset:], arena.buffer.memory[arena.buffer.offset:arena.buffer.offset+used])
	arena.buffer = newBuf
	// set offset to base + used
	arena.offset.Store(uintptr(newBuf.offset + used))
	return nil
}
