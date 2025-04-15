package memoryArena

import (
	"testing"
	"unsafe"
)

func TestMemoryArena(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	ptr, err := arena.Allocate(10)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if ptr == nil {
		t.Errorf("Error: ptr is nil")
	}

}

func TestMemoryArena_Reset(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	arena.Reset()

	for i := range arena.buffer.memory {
		if arena.buffer.memory[i] != 0 {
			t.Errorf("Error: memory is not reset")
		}
	}
}

func TestMemoryArena_Free(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	arena.Free()

	for i := range arena.buffer.memory {
		if arena.buffer.memory[i] != 0 {
			t.Errorf("Error: memory is not freed")
		}
	}
}

func TestMemoryArena_notEnoughSpace(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if arena.notEnoughSpace(100) {
		t.Errorf("Error: out of bounds")
	}
}

func TestMemoryArena_nextOffset(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if arena.nextOffset(10) != 10 {
		t.Errorf("Error: used capacity is not correct")
	}
}

func TestMemoryArena_alignOffset(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	arena.alignOffset(8)
	if arena.buffer.offset != 0 {
		t.Errorf("Error: offset is not aligned")
	}
}

func TestMemoryArenaBuffer_NewMemoryArenaBuffer(t *testing.T) {
	arena := NewMemoryArenaBuffer(100, 1)
	if arena.size != 100 {
		t.Errorf("Error: size is not correct")
	}
	if arena.offset != 0 {
		t.Errorf("Error: offset is not correct")
	}
}

func TestMemoryArenaResizePreserve(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	num, err := NewObject(arena, 5)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	arena.ResizePreserve(200)
	if arena.buffer.size != 200 {
		t.Errorf("Error: size is not preserved")
	}
	if *num != 5 {
		t.Errorf("Error: object is not preserved")
	}

}

func TestMemoryArenaResize(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	arena.Resize(200)
	if arena.buffer.size != 200 {
		t.Errorf("Error: size is not resized")
	}

}

func TestMemoryArenaAllocate(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	ptr, err := arena.Allocate(10)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if ptr == nil {
		t.Errorf("Error: ptr is nil")
	}

}

func TestMemoryArenaAllocateBuffer(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	ptr, err := arena.AllocateBuffer(10)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if ptr == nil {
		t.Errorf("Error: ptr is nil")
	}

}

func TestMemoryArena_AllocateNewValue(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	obj := 5
	ptr, err := arena.AllocateNewValue(10, obj)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if ptr == nil {
		t.Errorf("Error: ptr is nil")
	}
}

func TestMemoryArena_GetResult(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	ptr := arena.GetResult()
	if ptr == nil {
		t.Errorf("Error: ptr is nil")
	}
}

func TestMemoryArena_AllocateObject(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	obj := 5
	_, err = arena.AllocateObject(obj)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

func BenchmarkMemoryArena_AllocateObject(b *testing.B) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	obj := 5
	for i := 0; i < b.N; i++ {
		_, err = arena.AllocateObject(obj)
		if err != nil {
			b.Errorf("Error: %v", err)
		}
	}
}

func BenchmarkMemoryArena_AllocateNewValue(b *testing.B) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	obj := 5
	for i := 0; i < b.N; i++ {
		_, err = arena.AllocateNewValue(10, obj)
		if err != nil {
			b.Errorf("Error: %v", err)
		}
	}
}

func BenchmarkMemoryArena_AllocateBuffer(b *testing.B) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		_, err = arena.AllocateBuffer(10)
		if err != nil {
			b.Errorf("Error: %v", err)
		}
	}
}

func BenchmarkMemoryArena_Allocate(b *testing.B) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		_, err = arena.Allocate(10)
		if err != nil {
			b.Errorf("Error: %v", err)
		}
	}
}

func BenchmarkMemoryArena_Reset(b *testing.B) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		arena.Reset()
	}
}

func BenchmarkMemoryArena_Free(b *testing.B) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		arena.Free()
	}
}

func BenchmarkMemoryArena_GetResult(b *testing.B) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		arena.GetResult()
	}
}

func BenchmarkMemoryArena_ResizePreserve(b *testing.B) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		arena.ResizePreserve(200)
	}
}

func BenchmarkMemoryArena_Resize(b *testing.B) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		b.Errorf("Error: %v", err)
	}
	for i := 0; i < b.N; i++ {
		arena.Resize(200)
	}
}

// TestMemoryArenaBuffer_AlignmentCheck verifies that the effective base of the arena buffer is aligned.
func TestMemoryArenaBuffer_AlignmentCheck(t *testing.T) {
	// Test for a couple of alignment values.
	testCases := []struct {
		name      string
		alignment uintptr
	}{
		{"Alignment1", 1},
		{"Alignment8", 8},
		{"Alignment16", 16},
	}

	for _, tc := range testCases {
		// Create a new arena buffer with the given alignment.
		buffer := NewMemoryArenaBuffer(100, tc.alignment)
		basePtr := uintptr(unsafe.Pointer(&buffer.memory[0]))
		effectivePtr := basePtr + uintptr(buffer.offset)
		if effectivePtr%tc.alignment != 0 {
			t.Errorf("%s: effective pointer %#x is not aligned to %d bytes",
				tc.name, effectivePtr, tc.alignment)
		} else {
			t.Logf("%s: effective pointer %#x is correctly aligned to %d bytes",
				tc.name, effectivePtr, tc.alignment)
		}
	}
}

// TestMemoryArena_AllocationAlignment ensures that allocations return pointers that obey the type’s alignment.
func TestMemoryArena_AllocationAlignment(t *testing.T) {
	// Use int64 for which alignment is typically 8 bytes on 64-bit systems.
	arena, err := NewMemoryArena[int](10)
	if err != nil {
		t.Fatalf("Failed to create arena: %v", err)
	}

	ptr, err := arena.Allocate(8) // Allocate space for an int64.
	if err != nil {
		t.Fatalf("Allocation failed: %v", err)
	}

	effective := uintptr(ptr)
	requiredAlignment := uintptr(unsafe.Alignof(int64(0)))
	if effective%requiredAlignment != 0 {
		t.Errorf("Allocated pointer %#x is not aligned to %d bytes", effective, requiredAlignment)
	} else {
		t.Logf("Allocated pointer %#x is correctly aligned to %d bytes", effective, requiredAlignment)
	}
}

// TestMemoryArena_AllocateObjectAlignment checks that the pointer returned by AllocateObject
// is also aligned according to the type requirements.
func TestMemoryArena_AllocateObjectAlignment(t *testing.T) {
	type MyStruct struct {
		A int64
		B int32
	}
	arena, err := NewMemoryArena[MyStruct](100)
	if err != nil {
		t.Fatalf("Failed to create arena: %v", err)
	}

	obj := MyStruct{A: 42, B: 7}
	ptr, err := arena.AllocateObject(obj)
	if err != nil {
		t.Fatalf("AllocateObject failed: %v", err)
	}

	effective := uintptr(ptr)
	requiredAlignment := uintptr(unsafe.Alignof(MyStruct{}))
	if effective%requiredAlignment != 0 {
		t.Errorf("Object pointer %#x is not aligned to %d bytes", effective, requiredAlignment)
	} else {
		t.Logf("Object pointer %#x is correctly aligned to %d bytes", effective, requiredAlignment)
	}
}

// TestNewMemoryArenaBuffer_AlignmentOne verifies that when using an alignment of 1,
// no offset is needed.
func TestNewMemoryArenaBuffer_AlignmentOne(t *testing.T) {
	buf := NewMemoryArenaBuffer(100, 1)
	if buf.size != 100 {
		t.Errorf("expected size 100, got %d", buf.size)
	}
	if buf.offset != 0 {
		t.Errorf("expected offset 0 for alignment 1, got %d", buf.offset)
	}
	// Effective pointer check
	base := uintptr(unsafe.Pointer(&buf.memory[0]))
	effective := base + uintptr(buf.offset)
	if effective%1 != 0 {
		t.Errorf("effective pointer %#x is not aligned to 1 byte", effective)
	}
}

// TestNewMemoryArenaBuffer_LargeAlignment forces a misalignment branch.
// Using a large alignment (e.g. 4096) makes it very unlikely that the base is already aligned.
func TestNewMemoryArenaBuffer_LargeAlignment(t *testing.T) {
	const alignment = 4096
	buf := NewMemoryArenaBuffer(100, alignment)
	base := uintptr(unsafe.Pointer(&buf.memory[0]))
	remainder := base % alignment
	var expectedOffset int
	if remainder == 0 {
		// In the unlikely event that the base is aligned.
		expectedOffset = 0
	} else {
		expectedOffset = int(alignment - remainder)
	}
	if buf.offset != expectedOffset {
		t.Errorf("expected offset %d (base=0x%x, remainder=%d), got %d",
			expectedOffset, base, remainder, buf.offset)
	}
	// Ensure the effective pointer is aligned
	effective := base + uintptr(buf.offset)
	if effective%alignment != 0 {
		t.Errorf("effective pointer %#x is not aligned to %d bytes", effective, alignment)
	}
}

// TestNewMemoryArenaBuffer_AlternateAlignment tests a typical alignment (e.g. 8)
// and verifies that the computed offset is correct.
func TestNewMemoryArenaBuffer_AlternateAlignment(t *testing.T) {
	const alignment = 8
	buf := NewMemoryArenaBuffer(100, alignment)
	base := uintptr(unsafe.Pointer(&buf.memory[0]))
	remainder := base % alignment
	var expectedOffset int
	if remainder == 0 {
		expectedOffset = 0
	} else {
		expectedOffset = int(alignment - remainder)
	}
	if buf.offset != expectedOffset {
		t.Errorf("expected offset %d for alignment %d (base=0x%x, remainder=%d), got %d",
			expectedOffset, alignment, base, remainder, buf.offset)
	}
	effective := base + uintptr(buf.offset)
	if effective%alignment != 0 {
		t.Errorf("effective pointer %#x is not aligned to %d bytes", effective, alignment)
	}
}
