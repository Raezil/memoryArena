package memoryArena

import (
	"context"
	"testing"
	"unsafe"
)

func TestAllocateBeforeCancel(t *testing.T) {
	ctx := context.Background()
	ca, err := NewContextArena[int](ctx, 1024)
	if err != nil {
		t.Fatalf("failed to create ContextArena: %v", err)
	}
	ptr, err := ca.Allocate(8)
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}
	if ptr == nil {
		t.Fatal("Allocate returned nil pointer")
	}
}

func TestNewObjectBeforeCancel(t *testing.T) {
	ctx := context.Background()
	ca, err := NewContextArena[string](ctx, 1024)
	if err != nil {
		t.Fatalf("failed to create ContextArena: %v", err)
	}
	obj, err := ca.NewObject("hello")
	if err != nil {
		t.Fatalf("NewObject failed: %v", err)
	}
	if *obj != "hello" {
		t.Fatalf("expected 'hello', got '%s'", *obj)
	}
}

func TestAppendSliceBeforeCancel(t *testing.T) {
	ctx := context.Background()
	ca, err := NewContextArena[int](ctx, 1024)
	if err != nil {
		t.Fatalf("failed to create ContextArena: %v", err)
	}
	slice := []int{1, 2, 3}
	slice, err = ca.AppendSlice(slice, 4, 5)
	if err != nil {
		t.Fatalf("AppendSlice failed: %v", err)
	}
	if len(slice) != 5 || slice[3] != 4 || slice[4] != 5 {
		t.Fatalf("unexpected slice contents: %v", slice)
	}
}

func TestContextCancelResetsArena(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ca, err := NewContextArena[int](ctx, 16)
	if err != nil {
		t.Fatalf("failed to create ContextArena: %v", err)
	}
	// allocate to advance offset
	_, err = ca.Allocate(8)
	if err != nil {
		t.Fatalf("first Allocate failed: %v", err)
	}
	cancel()
	// after cancel, Allocate should error with context.Canceled
	_, err = ca.Allocate(8)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func BenchmarkContextArena_Allocate(b *testing.B) {
	ctx := context.Background()
	ca, _ := NewContextArena[int](ctx, b.N*8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ca.Allocate(8)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkContextArena_NewObject(b *testing.B) {
	ctx := context.Background()
	ca, _ := NewContextArena[int](ctx, b.N*int(unsafe.Sizeof(int(0))))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ca.NewObject(i)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkContextArena_AppendSlice(b *testing.B) {
	ctx := context.Background()
	// allocate 4x to accommodate doubling behavior
	ca, _ := NewContextArena[int](ctx, b.N*int(unsafe.Sizeof(int(0)))*4)
	slice := make([]int, 0, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var err error
		slice, err = ca.AppendSlice(slice, i)
		if err != nil {
			b.Fatal(err)
		}
	}
}
