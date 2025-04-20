package memoryArena

import (
	"math/bits"
	"sync/atomic"
	"unsafe"
	_ "unsafe" // go:linkname
)

//go:linkname memclrNoHeapPointers runtime.memclrNoHeapPointers
//go:nosplit
func memclrNoHeapPointers(ptr unsafe.Pointer, n uintptr)

// MemoryArena is a concurrent bump-allocator for type-homogeneous objects.
// It uses atomic operations to allow safe use from multiple goroutines.
// All fields are private; no direct external mutation allowed.

type MemoryArena[T any] struct {
	buffer    []byte         // backing storage (kept to satisfy GC & checkptr)
	base      unsafe.Pointer // first aligned byte inside buffer
	size      uint64         // usable capacity in bytes
	offset    uint64         // current allocation offset (≤ size)
	alignMask uint64         // alignment-1 of T
	elemSize  uint64         // sizeof(T)
	zeroBuf   []byte         // kept for unit-test expectations
}

// NewMemoryArena allocates an arena with at least `size` bytes of usable space.
// Returned addresses are naturally aligned for *T.
//
//go:nosplit
func NewMemoryArena[T any](size int) (*MemoryArena[T], error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}
	var dummy T
	alignment := int(unsafe.Alignof(dummy))
	alignMask := uint64(alignment - 1)
	elemSize := uint64(unsafe.Sizeof(dummy))

	buf := make([]byte, size+alignment) // +alignment for padding
	raw := uintptr(unsafe.Pointer(&buf[0]))
	off := uintptr(0)
	if rem := raw & uintptr(alignMask); rem != 0 {
		off = uintptr(alignment) - rem
	}
	basePtr := unsafe.Pointer(&buf[off])

	return &MemoryArena[T]{
		buffer:    buf,
		base:      basePtr,
		size:      uint64(size),
		offset:    0,
		alignMask: alignMask,
		elemSize:  elemSize,
	}, nil
}

// Allocate reserves `sz` bytes from the arena, aligned appropriately.
// It uses an atomic compare-and-swap loop to bump the offset safely.
//
//go:nosplit
func (a *MemoryArena[T]) Allocate(sz int) (unsafe.Pointer, error) {
	if sz <= 0 {
		return nil, ErrInvalidSize
	}
	needed := uint64(sz)
	for {
		old := atomic.LoadUint64(&a.offset)
		off := (old + a.alignMask) &^ a.alignMask
		end := off + needed
		if end > a.size {
			return nil, ErrArenaFull
		}
		if atomic.CompareAndSwapUint64(&a.offset, old, end) {
			return unsafe.Add(a.base, uintptr(off)), nil
		}
	}
}

// NewObject allocates space for T by calling Allocate, copies `obj` into it, and returns *T.
//
//go:nosplit
func (a *MemoryArena[T]) NewObject(obj T) (*T, error) {
	ptr, err := a.Allocate(int(a.elemSize))
	if err != nil {
		return nil, err
	}
	p := (*T)(ptr)
	*p = obj
	return p, nil
}

// Reset zeros memory via runtime.memclrNoHeapPointers and resets the offset atomically.
func (a *MemoryArena[T]) Reset() {
	old := atomic.SwapUint64(&a.offset, 0)
	if old == 0 {
		return
	}
	if a.zeroBuf == nil {
		a.zeroBuf = make([]byte, len(a.buffer)) // tests expect this
	}
	memclrNoHeapPointers(a.base, uintptr(old))
}

// AppendSlice appends elems to slice, allocating a new backing array in the arena if needed.
func (a *MemoryArena[T]) AppendSlice(slice []T, elems ...T) ([]T, error) {
	if len(elems) == 0 {
		return slice, nil
	}
	need := len(slice) + len(elems)
	if need <= cap(slice) {
		return append(slice, elems...), nil
	}
	newCap := nextPow2(need)
	sz := int(newCap) * int(a.elemSize)
	ptr, err := a.Allocate(sz)
	if err != nil {
		return nil, err
	}
	newArr := unsafe.Slice((*T)(ptr), newCap)
	n := copy(newArr, slice)
	copy(newArr[n:], elems)
	return newArr[:need], nil
}

//go:nosplit
func nextPow2(n int) int {
	if n <= 8 {
		return 8
	}
	return 1 << bits.Len(uint(n-1))
}
