package memoryArena

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

// Test basic Allocate, NewObject, Reset, and AppendSlice functionality
func TestAtomicArena_Basic(t *testing.T) {
	size := 1024
	arena, err := NewAtomicArena[int](size)
	if err != nil {
		t.Fatalf("failed to create arena: %v", err)
	}

	// Test Allocate within bounds
	ptr, err := arena.Allocate(8)
	if err != nil {
		t.Fatalf("Allocate error: %v", err)
	}
	// Write through pointer and read back
	p := (*int)(ptr)
	*p = 42
	if *p != 42 {
		t.Errorf("expected 42, got %d", *p)
	}

	// Test NewObject
	obj, err := arena.NewObject(99)
	if err != nil {
		t.Fatalf("NewObject error: %v", err)
	}
	if *obj != 99 {
		t.Errorf("expected NewObject to store 99, got %d", *obj)
	}

	// Test AppendSlice
	slice := []int{1, 2, 3}
	slice, err = arena.AppendSlice(slice, 4, 5, 6)
	if err != nil {
		t.Fatalf("AppendSlice error: %v", err)
	}
	expected := []int{1, 2, 3, 4, 5, 6}
	if !reflect.DeepEqual(slice, expected) {
		t.Errorf("AppendSlice result = %v, want %v", slice, expected)
	}

	// Test Reset zeros and allows reuse
	arena.Reset()
	// After reset, offset at zero; NewObject should allocate from start
	obj2, err := arena.NewObject(7)
	if err != nil {
		t.Fatalf("NewObject after reset error: %v", err)
	}
	if *obj2 != 7 {
		t.Errorf("expected NewObject after reset to store 7, got %d", *obj2)
	}
}

// Test concurrent Allocate and NewObject calls
func TestAtomicArena_Concurrent(t *testing.T) {
	n := 1000
	size := n * 8
	arena, err := NewAtomicArena[int](size)
	if err != nil {
		t.Fatalf("failed to create arena: %v", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			obj, err := arena.NewObject(val)
			if err != nil {
				errCh <- err
				return
			}
			if *obj != val {
				errCh <- fmt.Errorf("expected %d, got %d", val, *obj)
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
	for e := range errCh {
		t.Errorf("concurrent error: %v", e)
	}
}

// Benchmark NewObject across varying arena sizes
func BenchmarkAtomicArena_NewObject(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000, 1000000}
	for _, sz := range sizes {
		b.Run("Size_"+fmt.Sprint(sz), func(b *testing.B) {
			arena, _ := NewAtomicArena[int](sz)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = arena.NewObject(i)
			}
		})
	}
}

// Benchmark Allocate with fixed size per allocation
func BenchmarkAtomicArena_Allocate(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000, 1000000}
	sz := 8 // bytes per allocation
	for _, cap := range sizes {
		b.Run("Cap_"+fmt.Sprint(cap), func(b *testing.B) {
			arena, _ := NewAtomicArena[byte](cap)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = arena.Allocate(sz)
			}
		})
	}
}
