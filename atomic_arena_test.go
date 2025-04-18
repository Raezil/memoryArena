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

func TestAllocationAlignment(t *testing.T) {
	type S struct {
		A byte
		B int32
	}
	align := unsafe.Alignof(S{})
	a, _ := NewAtomicArena[S](int(align * 3))
	_, err := a.Allocate(1)
	if err != nil {
		t.Fatalf("unexpected error on initial misaligned allocate: %v", err)
	}
	basePtr := unsafe.Pointer(&a.buffer.memory[0])
	base := uintptr(basePtr)
	ptr, err := a.Allocate(int(unsafe.Sizeof(S{})))
	if err != nil {
		t.Fatalf("unexpected error on aligned allocate: %v", err)
	}
	off := uintptr(ptr) - base
	if off%align != 0 {
		t.Errorf("allocation not aligned: expected offset mod %d == 0, got %d", align, off%align)
	}
}

func TestAllocateNewValue_ErrArenaFull(t *testing.T) {
	size := int(unsafe.Sizeof(int(0))) - 1
	a, _ := NewAtomicArena[int](size)
	_, err := a.AllocateNewValue(1)
	if err != ErrArenaFull {
		t.Fatalf("expected ErrArenaFull, got %v", err)
	}
}

func TestAllocateNewValueAndAllocateObject(t *testing.T) {
	arena, err := NewAtomicArena[int](1024)
	if err != nil {
		t.Fatalf("failed to create arena: %v", err)
	}

	ptr, err := arena.AllocateNewValue(42)
	if err != nil {
		t.Fatalf("AllocateNewValue failed: %v", err)
	}
	val := *(*int)(ptr)
	if val != 42 {
		t.Errorf("expected value 42, got %d", val)
	}

	raw, err := arena.AllocateObject(100)
	if err != nil {
		t.Fatalf("AllocateObject failed: %v", err)
	}
	if *(*int)(raw) != 100 {
		t.Errorf("expected initialized int 100, got %d", *(*int)(raw))
	}

	*(*int)(raw) = 99
	if *(*int)(raw) != 99 {
		t.Errorf("expected stored value 99, got %d", *(*int)(raw))
	}

	if _, err := arena.AllocateObject("string"); err == nil {
		t.Error("expected error on type mismatch in AllocateObject, got nil")
	}
}

func TestAtomicReset(t *testing.T) {
	arena, err := NewAtomicArena[int](64)
	if err != nil {
		t.Fatalf("failed to create arena: %v", err)
	}
	ptr1, _ := arena.AllocateNewValue(7)
	ptr2, _ := arena.AllocateNewValue(8)
	*(*int)(ptr1) = 13
	*(*int)(ptr2) = 17

	arena.Reset()
	newPtr, err := arena.AllocateObject(0)
	if err != nil {
		t.Fatalf("AllocateObject after Reset failed: %v", err)
	}
	if *(*int)(newPtr) != 0 {
		t.Errorf("expected zero from fresh allocation after reset, got %d", *(*int)(newPtr))
	}
}

func TestConcurrentAllocations(t *testing.T) {
	arena, err := NewAtomicArena[uint64](1024 * 10)
	if err != nil {
		t.Fatalf("failed to create arena: %v", err)
	}
	var wg sync.WaitGroup
	count := 1000

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(val uint64) {
			defer wg.Done()
			ptr, err := arena.AllocateNewValue(val)
			if err != nil {
				t.Errorf("AllocateNewValue error in goroutine: %v", err)
				return
			}
			retrieved := *(*uint64)(ptr)
			if retrieved != val {
				t.Errorf("mismatched value: expected %d, got %d", val, retrieved)
			}
		}(uint64(i))
	}
	wg.Wait()
}

// --- Added tests for Resize and ResizePreserve ---

func TestResize_InvalidSize(t *testing.T) {
	a, _ := NewAtomicArena[int](16)
	err := a.Resize(0)
	if err != ErrInvalidSize {
		t.Fatalf("expected ErrInvalidSize, got %v", err)
	}
}

func TestResize_Functionality(t *testing.T) {
	a, _ := NewAtomicArena[int](4 * int(unsafe.Sizeof(int(0))))
	// exhaust initial arena
	for i := 0; i < 4; i++ {
		if _, err := a.Allocate(int(unsafe.Sizeof(int(0)))); err != nil {
			t.Fatalf("unexpected error before resize: %v", err)
		}
	}
	// should be full now
	if _, err := a.Allocate(1); err != ErrArenaFull {
		t.Fatalf("expected ErrArenaFull, got %v", err)
	}
	// resize to larger capacity
	err := a.Resize(8 * int(unsafe.Sizeof(int(0))))
	if err != nil {
		t.Fatalf("unexpected error on resize: %v", err)
	}
	// now can allocate up to new capacity
	for i := 0; i < 8; i++ {
		if _, err := a.Allocate(int(unsafe.Sizeof(int(0)))); err != nil {
			t.Fatalf("unexpected error after resize: %v", err)
		}
	}
	// should be full again
	if _, err := a.Allocate(1); err != ErrArenaFull {
		t.Fatalf("expected ErrArenaFull after resize allocations, got %v", err)
	}
}

func TestResizePreserve_InvalidSize(t *testing.T) {
	a, _ := NewAtomicArena[int](16)
	err := a.ResizePreserve(0)
	if err != ErrInvalidSize {
		t.Fatalf("expected ErrInvalidSize, got %v", err)
	}
}

func TestResizePreserve_TooSmall(t *testing.T) {
	a, _ := NewAtomicArena[byte](4)
	// use 2 bytes
	if _, err := a.Allocate(2); err != nil {
		t.Fatalf("setup allocate failed: %v", err)
	}
	err := a.ResizePreserve(1)
	if err != ErrNewSizeTooSmall {
		t.Fatalf("expected ErrNewSizeTooSmall, got %v", err)
	}
}

func TestResizePreserve_PreservesData(t *testing.T) {
	a, _ := NewAtomicArena[int](4 * int(unsafe.Sizeof(int(0))))
	// allocate two ints
	ptr1, _ := a.AllocateNewValue(11)
	ptr2, _ := a.AllocateNewValue(22)
	// ensure values are stored
	if *(*int)(ptr1) != 11 || *(*int)(ptr2) != 22 {
		t.Fatalf("values not stored before resize")
	}
	err := a.ResizePreserve(6 * int(unsafe.Sizeof(int(0))))
	if err != nil {
		t.Fatalf("unexpected error on ResizePreserve: %v", err)
	}
	// original pointers should still hold original values
	if *(*int)(ptr1) != 11 || *(*int)(ptr2) != 22 {
		t.Errorf("data not preserved after ResizePreserve: got %d, %d", *(*int)(ptr1), *(*int)(ptr2))
	}
	// allocate additional values up to new capacity
	for i := 0; i < 4; i++ {
		if _, err := a.Allocate(int(unsafe.Sizeof(int(0)))); err != nil {
			t.Fatalf("unexpected error allocating after preserve: %v", err)
		}
	}
	// should be full now
	if _, err := a.Allocate(1); err != ErrArenaFull {
		t.Fatalf("expected ErrArenaFull after preserve allocations, got %v", err)
	}
}
