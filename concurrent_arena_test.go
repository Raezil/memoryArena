// -----------------------------------------------------------------------------
//  Tests & Benchmarks – concurrent_arena_test.go
// -----------------------------------------------------------------------------

//go:build go1.22
// +build go1.22

package memoryArena

import (
	"fmt"
	"sync"
	"testing"
	"unsafe"
)

func TestConcurrentArena_ParallelAllocate(t *testing.T) {
	const arenaSize = 1 << 20 // 1 MiB
	ca, err := NewConcurrentArena[uint64](arenaSize)
	if err != nil {
		t.Fatalf("new arena: %v", err)
	}

	const goroutines = 8
	const perG = 10_000

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				if _, err := ca.NewObject(uint64(i)); err != nil {
					t.Errorf("alloc: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func BenchmarkConcurrentArenaNewObject(b *testing.B) {
	sizes := []int{100, 1_000, 10_000, 100_000, 1000000}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			ca, _ := NewConcurrentArena[int](n * 16)
			b.ResetTimer()
			b.ReportAllocs()
			b.SetBytes(int64(n * 8))
			for i := 0; i < b.N; i++ {
				for j := 0; j < n; j++ {
					if _, err := ca.NewObject(j); err != nil {
						b.Fatal(err)
					}
				}
				ca.Reset()
			}
		})
	}
}

func BenchmarkConcurrentArenaParallel(b *testing.B) {
	const arenaSize = 1 << 20 // 1 MiB
	ca, _ := NewConcurrentArena[int](arenaSize)
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ca.NewObject(42)
		}
	})
}

// -----------------------------------------------------------------------------
//  Unit tests – MemoryArena
// -----------------------------------------------------------------------------

func TestNewMemoryArena_Errors(t *testing.T) {
	if _, err := NewMemoryArena[int](0); err != ErrInvalidSize {
		t.Fatalf("want ErrInvalidSize, got %v", err)
	}
}

func TestMemoryArena_AllocateAndReset(t *testing.T) {
	const sz = 64
	a, _ := NewMemoryArena[byte](sz)

	// happy path allocate 16 bytes twice
	for i := 0; i < 2; i++ {
		p, _ := a.Allocate(16)
		// Fill and read back to ensure pointer is live
		b := unsafe.Slice((*byte)(p), 16)
		for i := range b {
			b[i] = 0xAA
		}
	}

	// Force arena full
	if _, err := a.Allocate(sz); err != ErrArenaFull {
		t.Fatalf("expected ErrArenaFull, got %v", err)
	}

	// Reset clears offset
	a.Reset()
	if a.offset != 0 {
		t.Fatalf("offset not reset: %d", a.offset)
	}
}

func TestMemoryArena_NewObject(t *testing.T) {
	a, _ := NewMemoryArena[int](128)
	const val = 42
	p, _ := a.NewObject(val)
	if *p != val {
		t.Fatalf("want %d, got %d", val, *p)
	}
}

func TestMemoryArena_AppendSlice(t *testing.T) {
	a, _ := NewMemoryArena[int](512)

	// Start with capacity 4
	s := make([]int, 0, 4)
	out, _ := a.AppendSlice(s, 1, 2, 3)
	if len(out) != 3 {
		t.Fatalf("len mismatch: %d", len(out))
	}

	// Force grow
	out, _ = a.AppendSlice(out, 4, 5, 6, 7)
	if cap(out) < 7 {
		t.Fatalf("capacity not grown: %d", cap(out))
	}
}

func TestNextPow2(t *testing.T) {
	cases := map[int]int{1: 8, 9: 16, 16: 16, 17: 32}
	for in, want := range cases {
		if got := nextPow2(in); got != want {
			t.Fatalf("nextPow2(%d) = %d, want %d", in, got, want)
		}
	}
}

// -----------------------------------------------------------------------------
//  Unit tests – ConcurrentArena
// -----------------------------------------------------------------------------

func TestConcurrentArena_Basic(t *testing.T) {
	const arenaSize = 1024
	ca, _ := NewConcurrentArena[int](arenaSize)
	_, _ = ca.NewObject(1)
	ca.Reset()
	if ca.arena.offset != 0 {
		t.Fatalf("offset not reset: %d", ca.arena.offset)
	}
}

func TestConcurrentArena_ParallelSafety(t *testing.T) {
	const arenaSize = 1 << 16 // 64 KiB
	ca, _ := NewConcurrentArena[int](arenaSize)

	var wg sync.WaitGroup
	const workers = 4
	const per = 2_000
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < per; i++ {
				if _, err := ca.NewObject(i + id*per); err != nil {
					t.Errorf("alloc failed: %v", err)
					return
				}
			}
		}(w)
	}
	wg.Wait()
}

// -----------------------------------------------------------------------------
//  Benchmarks – ensure they still build (coverage excluded)
// -----------------------------------------------------------------------------

func BenchmarkMemoryArenaNewObject(b *testing.B) {
	a, _ := NewMemoryArena[int](1 << 20)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		a.NewObject(i)
	}
}
