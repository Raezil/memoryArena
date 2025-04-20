package memoryArena

import (
	"sync/atomic"
	"unsafe"
	_ "unsafe" // for go:linkname
)

// AtomicArena is a concurrent bump‐allocator for type‐homogeneous objects.
// It uses atomic operations to allow safe allocations from multiple goroutines.
// Note: Reset is not concurrency‐safe and should be called when no allocations are in flight.

type AtomicArena[T any] struct {
	buffer    []byte         // backing storage (kept to satisfy GC & checkptr)
	base      unsafe.Pointer // first aligned byte inside buffer
	size      uintptr        // usable capacity in bytes
	alignMask uintptr        // alignment-1 of T
	elemSize  uintptr        // sizeof(T)
	offset    uint64         // current allocation offset in bytes (atomic)
	zeroBuf   []byte         // for unit‐test expectations
}

// NewAtomicArena allocates an arena with at least `size` bytes of usable space.
// Returned addresses are naturally aligned for *T.
func NewAtomicArena[T any](size int) (Arena[T], error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}
	var dummy T
	alignment := uintptr(unsafe.Alignof(dummy))
	alignMask := alignment - 1
	elemSize := uintptr(unsafe.Sizeof(dummy))

	buf := make([]byte, size+int(alignment))
	raw := uintptr(unsafe.Pointer(&buf[0]))
	off := uintptr(0)
	if rem := raw & alignMask; rem != 0 {
		off = alignment - rem
	}
	basePtr := unsafe.Pointer(&buf[off])

	return &AtomicArena[T]{
		buffer:    buf,
		base:      basePtr,
		size:      uintptr(size),
		alignMask: alignMask,
		elemSize:  elemSize,
		offset:    0,
	}, nil
}

// Allocate reserves sz bytes from the arena, aligned to T's alignment, returning a pointer.
func (a *AtomicArena[T]) Allocate(sz int) (unsafe.Pointer, error) {
	if sz <= 0 {
		return nil, ErrInvalidSize
	}
	szU := uintptr(sz)
	for {
		// load current offset
		head := atomic.LoadUint64(&a.offset)
		off0 := uintptr(head)
		// align up
		off := (off0 + a.alignMask) &^ a.alignMask
		end := off + szU
		// boundary check
		if end > a.size {
			return nil, ErrArenaFull
		}
		// try CAS
		newHead := uint64(end)
		if atomic.CompareAndSwapUint64(&a.offset, head, newHead) {
			// success
			return unsafe.Add(a.base, off), nil
		}
		// else retry
	}
}

// NewObject allocates space for T, copies obj into it, and returns *T.
func (a *AtomicArena[T]) NewObject(obj T) (*T, error) {
	ptr, err := a.Allocate(int(a.elemSize))
	if err != nil {
		return nil, err
	}
	r := (*T)(ptr)
	*r = obj
	return r, nil
}

// Reset zeros used memory and resets the offset to zero.
// Not safe to call concurrently with Allocate.
func (a *AtomicArena[T]) Reset() {
	head := atomic.LoadUint64(&a.offset)
	if head == 0 {
		return
	}
	if a.zeroBuf == nil {
		a.zeroBuf = make([]byte, len(a.buffer))
	}
	memclrNoHeapPointers(a.base, uintptr(head))
	atomic.StoreUint64(&a.offset, 0)
}

// AppendSlice appends elems to slice backed by this arena, resizing via the arena when needed.
func (a *AtomicArena[T]) AppendSlice(slice []T, elems ...T) ([]T, error) {
	if len(elems) == 0 {
		return slice, nil
	}
	sliceLen := len(slice)
	need := sliceLen + len(elems)
	if need <= cap(slice) {
		return append(slice, elems...), nil
	}
	newCap := nextPow2(need)
	sz := uintptr(newCap) * a.elemSize
	for {
		head := atomic.LoadUint64(&a.offset)
		off0 := uintptr(head)
		off := (off0 + a.alignMask) &^ a.alignMask
		end := off + sz
		if end > a.size {
			return nil, ErrArenaFull
		}
		if atomic.CompareAndSwapUint64(&a.offset, head, uint64(end)) {
			newArr := unsafe.Slice((*T)(unsafe.Add(a.base, off)), newCap)
			n := copy(newArr, slice)
			copy(newArr[n:], elems)
			return newArr[:need], nil
		}
	}
}

func (a *AtomicArena[T]) Offset() int {
	return int(a.offset)
}

func (a *AtomicArena[T]) Base() unsafe.Pointer {
	return a.base
}
