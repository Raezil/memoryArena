package memoryArena

import (
	"testing"
)

// TestNewAtomicMemoryArena_InvalidSize checks that creating an arena with non-positive size returns ErrInvalidSize.
func TestNewAtomicMemoryArena_InvalidSize(t *testing.T) {
	if _, err := NewAtomicMemoryArena[int](0); err != ErrInvalidSize {
		t.Errorf("expected ErrInvalidSize for size=0, got %v", err)
	}
	if _, err := NewAtomicMemoryArena[int](-10); err != ErrInvalidSize {
		t.Errorf("expected ErrInvalidSize for negative size, got %v", err)
	}
}

// TestAllocateAndReset verifies Allocate, ErrArenaFull, and Reset behavior.
func TestAllocateAndReset(t *testing.T) {
	a, err := NewAtomicMemoryArena[byte](64)
	if err != nil {
		t.Fatalf("unexpected error creating arena: %v", err)
	}

	// Allocate less than capacity
	p1, err := a.Allocate(16)
	if err != nil {
		t.Fatalf("unexpected error on Allocate: %v", err)
	}
	if p1 == nil {
		t.Errorf("Allocate returned nil pointer")
	}

	// Fill remaining capacity
	remaining := 64 - 16 - a.alignMask
	_, err = a.Allocate(remaining)
	if err != nil {
		t.Fatalf("unexpected error filling arena: %v", err)
	}

	// Next allocation should fail
	if _, err := a.Allocate(1); err != ErrArenaFull {
		t.Errorf("expected ErrArenaFull after capacity exceeded, got %v", err)
	}

	// Reset and allocate again
	a.Reset()
	p2, err := a.Allocate(8)
	if err != nil {
		t.Fatalf("unexpected error after Reset: %v", err)
	}
	if p2 == nil {
		t.Errorf("Allocate returned nil pointer after Reset")
	}
}

// TestNewObject ensures NewObject stores the correct value.
func TestNewObject(t *testing.T) {
	a, err := NewAtomicMemoryArena[int](1024)
	if err != nil {
		t.Fatalf("unexpected error creating arena: %v", err)
	}

	val := 42
	ptr, err := a.NewObject(val)
	if err != nil {
		t.Fatalf("NewObject error: %v", err)
	}
	if *ptr != val {
		t.Errorf("expected stored value %d, got %d", val, *ptr)
	}
}

// TestAppendSlice verifies that AppendSlice returns correct slice contents.
func TestAppendSlice(t *testing.T) {
	a, err := NewAtomicMemoryArena[int](128)
	if err != nil {
		t.Fatalf("unexpected error creating arena: %v", err)
	}

	slice := make([]int, 0, 2)
	slice, err = a.AppendSlice(slice, 1, 2, 3)
	if err != nil {
		t.Fatalf("AppendSlice error: %v", err)
	}
	if len(slice) != 3 {
		t.Errorf("expected length 3, got %d", len(slice))
	}
	for i, v := range []int{1, 2, 3} {
		if slice[i] != v {
			t.Errorf("expected slice[%d]==%d, got %d", i, v, slice[i])
		}
	}
}

// TestConcurrentAllocate ensures concurrent Allocate calls succeed uniquely.
func TestConcurrentAllocate(t *testing.T) {
	const count = 100
	a, err := NewAtomicMemoryArena[byte](count)
	if err != nil {
		t.Fatalf("unexpected error creating arena: %v", err)
	}

	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		go func() {
			_, err := a.Allocate(1)
			errs <- err
		}()
	}
	for i := 0; i < count; i++ {
		if err := <-errs; err != nil {
			t.Errorf("Allocate failed concurrently: %v", err)
		}
	}
	// Further allocation should fail
	if _, err := a.Allocate(1); err != ErrArenaFull {
		t.Errorf("expected ErrArenaFull after concurrent full, got %v", err)
	}
}

// BenchmarkAllocate measures the speed of Allocate.
func BenchmarkAllocate(b *testing.B) {
	a, _ := NewAtomicMemoryArena[byte](b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Allocate(1)
	}
}

// BenchmarkNewObject measures the speed of NewObject.
func BenchmarkNewObject(b *testing.B) {
	a, _ := NewAtomicMemoryArena[int](b.N * 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.NewObject(i)
	}
}
