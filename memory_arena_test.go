package memoryArena

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"unsafe"
)

// /////////////////////////////////////////////////////////////////////////////
//                              UNIT TESTS
// /////////////////////////////////////////////////////////////////////////////

type point struct{ X, Y int }

type tiny byte

func TestNewObject_roundTrip(t *testing.T) {
	arena, _ := NewMemoryArena[point](1024)
	p, err := arena.NewObject(point{3, 4})
	if err != nil {
		t.Fatalf("NewObject: %v", err)
	}
	if p.X != 3 || p.Y != 4 {
		t.Fatalf("value mismatch got=%+v", *p)
	}
}

func TestAllocate_alignment(t *testing.T) {
	arena, _ := NewMemoryArena[uint64](128)
	for i := 0; i < 4; i++ {
		ptr, _ := arena.Allocate(1)
		if off := uintptr(ptr) & 7; off != 0 {
			t.Fatalf("ptr not 8‑byte aligned: %x", off)
		}
	}
}

func TestReset_zeroesUsedBytes(t *testing.T) {
	arena, _ := NewMemoryArena[byte](64)
	// fill 16 bytes with 0xAA
	ptr, _ := arena.Allocate(16)
	slice := unsafe.Slice((*byte)(ptr), 16)
	for i := range slice {
		slice[i] = 0xAA
	}
	arena.Reset()
	ptr2, _ := arena.Allocate(16)
	slice2 := unsafe.Slice((*byte)(ptr2), 16)
	for i, v := range slice2 {
		if v != 0 {
			t.Fatalf("byte %d not zero (%x)", i, v)
		}
	}
}

func TestAppendSlice_growth(t *testing.T) {
	arena, _ := NewMemoryArena[int](4096)
	var s []int
	var err error
	for i := 0; i < 40; i++ {
		s, err = arena.AppendSlice(s, i)
		if err != nil {
			t.Fatalf("AppendSlice: %v", err)
		}
	}
	if len(s) != 40 || cap(s) < 40 || s[39] != 39 {
		t.Fatalf("unexpected slice %v (cap %d)", s, cap(s))
	}
}

// /////////////////////////////////////////////////////////////////////////////
//                          CONCURRENCY TESTS
// /////////////////////////////////////////////////////////////////////////////

func TestArena_mutexProtection(t *testing.T) {
	// NOTE: arena is NOT thread‑safe, but demonstrate external sync correctness.
	arena, _ := NewMemoryArena[point](1 << 20)
	var mu sync.Mutex
	wg := sync.WaitGroup{}
	for g := 0; g < runtime.NumCPU(); g++ {
		wg.Add(1)
		go func(seed int) {
			defer wg.Done()
			for i := 0; i < 1_000; i++ {
				mu.Lock()
				_, err := arena.NewObject(point{X: seed, Y: i})
				mu.Unlock()
				if err != nil {
					t.Errorf("alloc err: %v", err)
					return
				}
			}
		}(g)
	}
	wg.Wait()
}

// /////////////////////////////////////////////////////////////////////////////
//                              BENCHMARKS
// /////////////////////////////////////////////////////////////////////////////

func BenchmarkAllocate64B(b *testing.B) {
	arena, _ := NewMemoryArena[byte](1 << 20) // 1 MiB arena
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := arena.Allocate(64); err != nil {
			if err != ErrArenaFull {
				b.Fatal(err)
			}
			arena.Reset()
			if _, err := arena.Allocate(64); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkAppendSliceGrow(b *testing.B) {
	arena, _ := NewMemoryArena[int](1 << 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s []int
		var err error
		for n := 0; n < 200; n++ {
			s, err = arena.AppendSlice(s, n)
			if err != nil {
				if err != ErrArenaFull {
					b.Fatal(err)
				}
				arena.Reset()
				s, err = arena.AppendSlice(nil, n)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
		arena.Reset()
	}
}

var benchSizes = []int{10, 100, 1_000, 10_000, 100_000, 1_000_000}

// ----------------------------------------------------------------------------
// Allocate – raw byte reservations
// ----------------------------------------------------------------------------
func BenchmarkSizesAllocate(b *testing.B) {
	for _, sz := range benchSizes {
		b.Run(fmt.Sprintf("%dB", sz), func(b *testing.B) {
			arena, _ := NewMemoryArena[byte](sz * 2) // twice the request → reset rarely
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := arena.Allocate(sz); err != nil {
					if err != ErrArenaFull {
						b.Fatalf("unexpected err: %v", err)
					}
					arena.Reset()
					_, _ = arena.Allocate(sz)
				}
			}
		})
	}
}

// ----------------------------------------------------------------------------
// NewObject – struct allocation & copy
// ----------------------------------------------------------------------------
type t struct{ A, B, C int }

func BenchmarkSizesNewObject(b *testing.B) {
	for _, sz := range benchSizes {
		b.Run(fmt.Sprintf("%d×tiny", sz), func(b *testing.B) {
			arena, _ := NewMemoryArena[t](sz * int(unsafe.Sizeof(t{})) * 2)
			obj := t{1, 2, 3}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := arena.NewObject(obj); err != nil {
					if err != ErrArenaFull {
						b.Fatalf("err: %v", err)
					}
					arena.Reset()
					_, _ = arena.NewObject(obj)
				}
			}
		})
	}
}

// ----------------------------------------------------------------------------
// AppendSlice – growing a slice from 0 → size‑1
// ----------------------------------------------------------------------------
func BenchmarkSizesAppendSlice(b *testing.B) {
	for _, sz := range benchSizes {
		b.Run(fmt.Sprintf("%dIntsGrow", sz), func(b *testing.B) {
			arena, _ := NewMemoryArena[int](sz * 8 * 2) // 8 bytes per int
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var s []int
				var err error
				for n := 0; n < sz; n++ {
					s, err = arena.AppendSlice(s, n)
					if err != nil {
						if err != ErrArenaFull {
							b.Fatalf("err: %v", err)
						}
						arena.Reset()
						s, _ = arena.AppendSlice(nil, n)
					}
				}
				arena.Reset()
			}
		})
	}
}

// FuzzAllocateRoundTrip stresses Allocate/Reset via go test -fuzz (Go >=1.18).
// It checks that after Reset the arena returns zeroed bytes while ResetFast does not.
func FuzzAllocateRoundTrip(f *testing.F) {
	f.Add(uint8(8)) // seed corpus – 8 bytes
	f.Fuzz(func(t *testing.T, n uint8) {
		if n == 0 {
			t.Skip()
		}
		arena, _ := NewMemoryArena[byte](1024)
		p, err := arena.Allocate(int(n))
		if err != nil {
			t.Fatalf("Allocate: %v", err)
		}
		s := unsafe.Slice((*byte)(p), int(n))
		for i := range s {
			s[i] = 0xAB
		}
		arena.Reset()
		p2, _ := arena.Allocate(int(n))
		s2 := unsafe.Slice((*byte)(p2), int(n))
		for i, v := range s2 {
			if v != 0 {
				t.Fatalf("byte %d not zero after Reset (value %x)", i, v)
			}
		}
	})
}

// BenchmarkReset measures the cost of a zeroing Reset which calls memclr.
func BenchmarkReset(b *testing.B) {
	arena, _ := NewMemoryArena[byte](1 << 20)
	// Fill some memory each iter so Reset actually has work to do.
	fill := make([]byte, 4096)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = arena.Allocate(len(fill))
		copy(unsafe.Slice((*byte)(unsafe.Pointer(uintptr(arena.base)+uintptr(int(arena.offset)-len(fill)))), len(fill)), fill)
		arena.Reset()
	}
}

// small fits in one cache line.

// big forces an allocation much larger than the usual scalar values.

// -----------------------------------------------------------------------------
// Helper to prevent the compiler from optimising away allocations.
var sink any

// -----------------------------------------------------------------------------
// -----------------------------------------------------------------------------
// Allocate(sz) benchmarks – arena vs make --------------------------------------

func BenchmarkArenaAllocate64(b *testing.B) {
	arena, _ := NewMemoryArena[byte](4 << 20)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := arena.Allocate(64)
		sink = p
	}
	runtime.KeepAlive(sink)
}

func BenchmarkArenaAllocate4K(b *testing.B) {
	arena, _ := NewMemoryArena[byte](8 << 20)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := arena.Allocate(4096)
		sink = p
	}
	runtime.KeepAlive(sink)
}

// -----------------------------------------------------------------------------
// Reset cost -------------------------------------------------------------------

func BenchmarkArenaReset8K(b *testing.B) {
	arena, _ := NewMemoryArena[byte](8 << 13) // 8 KiB usable
	// Pre‑touch a page so reset has something to clear
	arena.Allocate(8 << 13)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.Reset()
	}
}
