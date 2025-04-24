package memoryArena

import "testing"

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
