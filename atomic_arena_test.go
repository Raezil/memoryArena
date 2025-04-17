package memoryArena

import (
	"sync"
	"testing"
	"unsafe"
)

func TestNewAtomicArena_InvalidSize(t *testing.T) {
	_, err := NewAtomicArena[int](0)
	if err != ErrInvalidSize {
		t.Fatalf("expected ErrInvalidSize, got %v", err)
	}
}

func TestAllocate_InvalidSize(t *testing.T) {
	a, _ := NewAtomicArena[int](1024)
	_, err := a.Allocate(0)
	if err != ErrInvalidSize {
		t.Fatalf("expected ErrInvalidSize, got %v", err)
	}
}

func TestAllocateAndExhaust(t *testing.T) {
	size := 16
	a, _ := NewAtomicArena[byte](size)
	var ptrs []unsafe.Pointer
	for {
		ptr, err := a.Allocate(1)
		if err == ErrArenaFull {
			break
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ptrs = append(ptrs, ptr)
	}
	if len(ptrs) == 0 {
		t.Fatal("expected at least one allocation")
	}
}

func TestAllocateNewValue(t *testing.T) {
	a, _ := NewAtomicArena[int](128)
	v := 42
	ptr, err := a.AllocateNewValue(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := *(*int)(ptr)
	if got != v {
		t.Fatalf("expected %d, got %d", v, got)
	}
}

func TestReset(t *testing.T) {
	size := 10
	a, _ := NewAtomicArena[byte](size)
	for i := 0; i < size; i++ {
		if _, err := a.Allocate(1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	_, err := a.Allocate(1)
	if err != ErrArenaFull {
		t.Fatalf("expected ErrArenaFull before reset, got %v", err)
	}
	a.Reset()
	_, err = a.Allocate(1)
	if err != nil {
		t.Fatalf("unexpected error after reset: %v", err)
	}
}

func TestConcurrentAllocate(t *testing.T) {
	total := 1000
	// capacity in bytes: total * size of int
	a, _ := NewAtomicArena[int](total * int(unsafe.Sizeof(int(0))))
	var wg sync.WaitGroup
	ptrs := make([]unsafe.Pointer, total)
	wg.Add(total)
	for i := 0; i < total; i++ {
		go func(idx, val int) {
			defer wg.Done()
			ptr, err := a.AllocateNewValue(val)
			if err != nil {
				t.Errorf("allocation error: %v", err)
				return
			}
			ptrs[idx] = ptr
		}(i, i)
	}
	wg.Wait()
	seen := make(map[uintptr]bool)
	for _, p := range ptrs {
		up := uintptr(p)
		if seen[up] {
			t.Errorf("duplicate pointer: %v", p)
		}
		seen[up] = true
	}
}

// Test that allocations respect type alignment, covering the unaligned offset path.
func TestAllocationAlignment(t *testing.T) {
	type S struct {
		A byte
		B int32
	}
	align := unsafe.Alignof(S{})
	// capacity enough for a misaligned and aligned allocation
	a, _ := NewAtomicArena[S](int(align * 3))
	// first allocate 1 byte to force a misalignment
	_, err := a.Allocate(1)
	if err != nil {
		t.Fatalf("unexpected error on initial misaligned allocate: %v", err)
	}
	// record base address
	basePtr := unsafe.Pointer(&a.buffer.memory[0])
	base := uintptr(basePtr)
	// allocate a full S to trigger alignment adjustment
	ptr, err := a.Allocate(int(unsafe.Sizeof(S{})))
	if err != nil {
		t.Fatalf("unexpected error on aligned allocate: %v", err)
	}
	off := uintptr(ptr) - base
	if off%align != 0 {
		t.Errorf("allocation not aligned: expected offset mod %d == 0, got %d", align, off%align)
	}
}

// Test AllocateNewValue error propagation when arena is full
func TestAllocateNewValue_ErrArenaFull(t *testing.T) {
	size := int(unsafe.Sizeof(int(0))) - 1
	a, _ := NewAtomicArena[int](size)
	_, err := a.AllocateNewValue(1)
	if err != ErrArenaFull {
		t.Fatalf("expected ErrArenaFull, got %v", err)
	}
}
