package memoryArena

import (
	"sync/atomic"
	"unsafe"
	_ "unsafe" // for go:linkname
)

// AtomicMemoryArena is a thread-safe bump-allocator for type-homogeneous objects.
// All fields are private; no direct external mutation allowed.
// It uses atomic operations to allow concurrent allocations.
type AtomicMemoryArena[T any] struct {
	buffer    []byte         // backing storage (kept to satisfy GC & checkptr)
	base      unsafe.Pointer // first aligned byte inside buffer
	size      int            // usable capacity in bytes
	offset    atomic.Int64   // current allocation offset (≤ size)
	alignMask int            // alignment-1 of T
	elemSize  int            // sizeof(T)
	zeroBuf   []byte         // kept for tests
}

// NewAtomicMemoryArena allocates an arena with at least `size` bytes of usable space, thread-safe.
func NewAtomicMemoryArena[T any](size int) (*AtomicMemoryArena[T], error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}
	var dummy T
	alignment := int(unsafe.Alignof(dummy))
	alignMask := alignment - 1
	elemSize := int(unsafe.Sizeof(dummy))

	buf := make([]byte, size+alignment)
	raw := uintptr(unsafe.Pointer(&buf[0]))
	off := 0
	if rem := int(raw) & alignMask; rem != 0 {
		off = alignment - rem
	}
	basePtr := unsafe.Pointer(&buf[off])

	a := &AtomicMemoryArena[T]{
		buffer:    buf,
		base:      basePtr,
		size:      size,
		alignMask: alignMask,
		elemSize:  elemSize,
	}
	// offset starts at 0
	return a, nil
}

// Allocate reserves sz bytes and returns a pointer to the block, or ErrArenaFull.
func (a *AtomicMemoryArena[T]) Allocate(sz int) (unsafe.Pointer, error) {
	if sz <= 0 {
		return nil, ErrInvalidSize
	}
	for {
		old := a.offset.Load()
		off := (int(old) + a.alignMask) &^ a.alignMask
		newOff := off + sz
		if newOff > a.size {
			return nil, ErrArenaFull
		}
		if a.offset.CompareAndSwap(old, int64(newOff)) {
			return unsafe.Add(a.base, uintptr(off)), nil
		}
	}
}

// NewObject allocates space for one element of T, stores obj, and returns its pointer.
func (a *AtomicMemoryArena[T]) NewObject(obj T) (*T, error) {
	for {
		old := a.offset.Load()
		off := (int(old) + a.alignMask) &^ a.alignMask
		newOff := off + a.elemSize
		if newOff > a.size {
			return nil, ErrArenaFull
		}
		if a.offset.CompareAndSwap(old, int64(newOff)) {
			ptr := (*T)(unsafe.Add(a.base, uintptr(off)))
			*ptr = obj
			return ptr, nil
		}
	}
}

// AppendSlice appends elems to slice, using arena-backed storage when capacity exceeded.
func (a *AtomicMemoryArena[T]) AppendSlice(slice []T, elems ...T) ([]T, error) {
	if len(elems) == 0 {
		return slice, nil
	}
	need := len(slice) + len(elems)
	if need <= cap(slice) {
		return append(slice, elems...), nil
	}
	newCap := nextPow2(need)
	sz := newCap * a.elemSize
	for {
		old := a.offset.Load()
		off := (int(old) + a.alignMask) &^ a.alignMask
		newOff := off + sz
		if newOff > a.size {
			return nil, ErrArenaFull
		}
		if a.offset.CompareAndSwap(old, int64(newOff)) {
			newArr := unsafe.Slice((*T)(unsafe.Add(a.base, uintptr(off))), newCap)
			n := copy(newArr, slice)
			copy(newArr[n:], elems)
			return newArr[:need], nil
		}
	}
}

// Reset clears all allocated data, setting offset back to zero (not thread-safe).
func (a *AtomicMemoryArena[T]) Reset() {
	old := a.offset.Load()
	if old == 0 {
		return
	}
	if a.zeroBuf == nil {
		a.zeroBuf = make([]byte, len(a.buffer))
	}
	memclrNoHeapPointers(a.base, uintptr(old))
	a.offset.Store(0)
}
