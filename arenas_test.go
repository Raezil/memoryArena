package memoryArena

import (
	"runtime"
	"sync"
	"testing"
)

type Obj100 [100]byte
type Obj1000 [1000]byte
type Obj10000 [10000]byte
type Obj100000 [100000]byte
type Obj1000000 [1000000]byte
type Obj10000000 [10000000]byte
type Obj100000000 [100000000]byte

func makeBench(size int) func(b *testing.B) {
	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = make([]byte, size)
		}
	}
}

func BenchmarkNativeMakeSlice100(b *testing.B)       { makeBench(100)(b) }
func BenchmarkNativeMakeSlice1000(b *testing.B)      { makeBench(1000)(b) }
func BenchmarkNativeMakeSlice10000(b *testing.B)     { makeBench(10000)(b) }
func BenchmarkNativeMakeSlice100000(b *testing.B)    { makeBench(100000)(b) }
func BenchmarkNativeMakeSlice1000000(b *testing.B)   { makeBench(1000000)(b) }
func BenchmarkNativeMakeSlice10000000(b *testing.B)  { makeBench(10000000)(b) }
func BenchmarkNativeMakeSlice100000000(b *testing.B) { makeBench(100000000)(b) }

// --- Native new([N]byte) benchmarks ---

func BenchmarkNative_NewObject100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = new(Obj100)
	}
}

func BenchmarkNative_NewObject1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = new(Obj1000)
	}
}

func BenchmarkNative_NewObject10000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = new(Obj10000)
	}
}

func BenchmarkNative_NewObject100000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = new(Obj100000)
	}
}

func BenchmarkNative_NewObject1000000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = new(Obj1000000)
	}
}

func BenchmarkNative_NewObject10000000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = new(Obj10000000)
	}
}

func BenchmarkNative_NewObject100000000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = new(Obj100000000)
	}
}

// Dummy is a 1KB object used for arena tests.

// Dummy is a 1KB object used for arena tests.
type Dummy struct{ data [1024]byte }

// TestMemoryArenaMemoryLeak allocates until ErrArenaFull, resets, and checks for leaks.
func TestMemoryArenaMemoryLeak(t *testing.T) {
	arena, err := NewMemoryArena[Dummy](10 * 1024 * 1024)
	if err != nil {
		t.Fatal(err)
	}

	// Drain arena until full
	for i := 0; ; i++ {
		if _, err := arena.NewObject(Dummy{}); err != nil {
			if err == ErrArenaFull {
				break
			}
			t.Fatalf("unexpected error at allocation %d: %v", i, err)
		}
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocBefore := m.Alloc

	arena.Reset()
	runtime.GC()
	runtime.ReadMemStats(&m)
	allocAfter := m.Alloc

	t.Logf("MemoryArena: Alloc before reset: %d, after reset: %d", allocBefore, allocAfter)
	if allocAfter > allocBefore+1024*1024 {
		t.Errorf("MemoryArena potential leak: delta %d exceeds threshold", allocAfter-allocBefore)
	}
}

// TestAtomicArenaMemoryLeak allocates until ErrArenaFull, resets, and checks for leaks.
func TestAtomicArenaMemoryLeak(t *testing.T) {
	arena, err := NewAtomicArena[Dummy](10 * 1024 * 1024)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; ; i++ {
		if _, err := arena.NewObject(Dummy{}); err != nil {
			if err == ErrArenaFull {
				break
			}
			t.Fatalf("unexpected error at allocation %d: %v", i, err)
		}
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocBefore := m.Alloc

	arena.Reset()
	runtime.GC()
	runtime.ReadMemStats(&m)
	allocAfter := m.Alloc

	t.Logf("AtomicArena: Alloc before reset: %d, after reset: %d", allocBefore, allocAfter)
	if allocAfter > allocBefore+1024*1024 {
		t.Errorf("AtomicArena potential leak: delta %d exceeds threshold", allocAfter-allocBefore)
	}
}

// TestConcurrentArenaMemoryLeak performs concurrent allocations under capacity, resets, and checks for leaks.
func TestConcurrentArenaMemoryLeak(t *testing.T) {
	arena, err := NewConcurrentArena[Dummy](10 * 1024 * 1024)
	if err != nil {
		t.Fatal(err)
	}

	const workers = 100
	const opsPerWorker = 100 // total = 10000 < capacity ~10240

	var (
		wg    sync.WaitGroup
		errCh = make(chan error, 1)
	)

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerWorker; j++ {
				if _, err := arena.NewObject(Dummy{}); err != nil {
					switch err {
					case ErrArenaFull:
						t.Errorf("worker %d: unexpected ErrArenaFull at op %d", id, j)
					default:
						select {
						case errCh <- err:
						default:
						}
					}
					return
				}
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	if err, ok := <-errCh; ok {
		t.Fatal(err)
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocBefore := m.Alloc

	arena.Reset()
	runtime.GC()
	runtime.ReadMemStats(&m)
	allocAfter := m.Alloc

	t.Logf("ConcurrentArena: Alloc before reset: %d, after reset: %d", allocBefore, allocAfter)
	if allocAfter > allocBefore+1024*1024 {
		t.Errorf("ConcurrentArena potential leak: delta %d exceeds threshold", allocAfter-allocBefore)
	}
}
