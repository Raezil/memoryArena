package memoryArena

import (
	"context"
	"fmt"
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

// BenchmarkContextArena_Allocate measures per-op cost of Allocate(sz) for various sizes.
func BenchmarkContextArena_Allocates(b *testing.B) {
	sizes := []int{1, 10, 100, 1000, 10000, 100000, 1000000}
	const capBytes = 1 << 30 // 1 GiB
	ctx := context.Background()

	for _, sz := range sizes {
		sz := sz
		b.Run(fmt.Sprintf("Allocate_%dB", sz), func(b *testing.B) {
			arena, err := NewContextArena[byte](ctx, capBytes)
			if err != nil {
				b.Fatalf("failed to create context arena: %v", err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := arena.Allocate(sz)
				if err != nil {
					b.Fatalf("allocate failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkContextArena_NewObject measures per-op cost of NewObject(obj) for various object sizes.
func BenchmarkContextArena_NewObjects(b *testing.B) {
	const capBytes = 1 << 30 // 1 GiB

	types := []struct {
		name string
		sz   int
	}{{"1B", 1}, {"10B", 10}, {"100B", 100}, {"1KiB", 1024}, {"10KiB", 10 * 1024}, {"100KiB", 100 * 1024}, {"1MiB", 1024 * 1024}}

	for _, tt := range types {
		tt := tt
		b.Run(fmt.Sprintf("NewObject_%s", tt.name), func(b *testing.B) {
			// define an array type of the given size
			type T [] /* placeholder */ byte
			switch tt.sz {
			case 1:
				type T [1]byte
			case 10:
				type T [10]byte
			case 100:
				type T [100]byte
			case 1024:
				type T [1024]byte
			case 10 * 1024:
				type T [10240]byte
			case 100 * 1024:
				type T [102400]byte
			case 1024 * 1024:
				type T [1048576]byte
			}
			ctx := context.Background()

			arena, err := NewContextArena[T](ctx, capBytes)
			if err != nil {
				b.Fatalf("failed to create context arena: %v", err)
			}
			var obj T
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := arena.NewObject(obj)
				if err != nil {
					b.Fatalf("new object failed: %v", err)
				}
			}
		})
	}
}
