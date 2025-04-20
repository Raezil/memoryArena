package memoryArena

import (
	"context"
	"testing"
	"unsafe"
)

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

// Benchmarks for NewObject with varying object sizes
func BenchmarkContextArena_NewObject_100B(b *testing.B) {
	ctx := context.Background()
	n := int(unsafe.Sizeof([100]byte{}))
	ca, _ := NewContextArena[[100]byte](ctx, b.N*n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ca.NewObject([100]byte{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkContextArena_NewObject_1KB(b *testing.B) {
	ctx := context.Background()
	n := int(unsafe.Sizeof([1000]byte{}))
	ca, _ := NewContextArena[[1000]byte](ctx, b.N*n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ca.NewObject([1000]byte{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkContextArena_NewObject_10KB(b *testing.B) {
	ctx := context.Background()
	n := int(unsafe.Sizeof([10000]byte{}))
	ca, _ := NewContextArena[[10000]byte](ctx, b.N*n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ca.NewObject([10000]byte{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkContextArena_NewObject_100KB(b *testing.B) {
	ctx := context.Background()
	n := int(unsafe.Sizeof([100000]byte{}))
	ca, _ := NewContextArena[[100000]byte](ctx, b.N*n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ca.NewObject([100000]byte{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkContextArena_NewObject_1MB(b *testing.B) {
	ctx := context.Background()
	n := int(unsafe.Sizeof([1000000]byte{}))
	ca, _ := NewContextArena[[1000000]byte](ctx, b.N*n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ca.NewObject([1000000]byte{})
		if err != nil {
			b.Fatal(err)
		}
	}
}
