// memory_arena.go
package memoryArena

import (
	"unsafe"
)

// MemoryArenaBuffer represents a contiguous block of memory for allocations.
type MemoryArenaBuffer struct {
	memory []byte
	size   int
	offset int
}

// NewMemoryArenaBuffer creates a new buffer of the given size with the specified alignment.
func NewMemoryArenaBuffer(size int, alignment uintptr) *MemoryArenaBuffer {
	// Allocate extra space to ensure we can align the base.
	buf := make([]byte, size+int(alignment))

	// Audited: converting &buf[0] to uintptr is safe because buf has length >= 1
	base := uintptr(unsafe.Pointer(&buf[0])) // #nosec G103
	offset := 0
	if rem := base % alignment; rem != 0 {
		offset = int(alignment - rem)
	}

	return &MemoryArenaBuffer{
		memory: buf,
		size:   size,
		offset: offset,
	}
}

// MemoryArena provides bump allocation for objects of type T.
type MemoryArena[T any] struct {
	buffer MemoryArenaBuffer
}

// NewMemoryArena creates a new MemoryArena with the specified capacity.
// Returns ErrInvalidSize if size <= 0.
func NewMemoryArena[T any](size int) (*MemoryArena[T], error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}
	alignment := unsafe.Alignof(*new(T))
	buffer := NewMemoryArenaBuffer(size, alignment)
	return &MemoryArena[T]{buffer: *buffer}, nil
}

func ptrAt(mem []byte, offset int) unsafe.Pointer {
	base := unsafe.Pointer(&mem[0])                        // #nosec G103
	return unsafe.Pointer(uintptr(base) + uintptr(offset)) // #nosec G103
}

// GetResult returns a pointer to the current offset in the buffer.
func (arena *MemoryArena[T]) GetResult() unsafe.Pointer {
	// Audited: compute pointer from base to avoid slizce-bound checks and ensure correct alignment
	return ptrAt(arena.buffer.memory, arena.buffer.offset) // #nosec G103
}

// Allocate reserves size bytes and returns a pointer or ErrInvalidSize/ErrArenaFull.
func (arena *MemoryArena[T]) Allocate(size int) (unsafe.Pointer, error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}
	return arena.AllocateBuffer(size)
}

func (arena *MemoryArena[T]) AllocateBuffer(size int) (unsafe.Pointer, error) {
	alignment := unsafe.Alignof(*new(T))
	arena.alignOffset(alignment)
	if arena.nextOffset(size) > arena.buffer.size {
		return nil, ErrArenaFull
	}
	ptr := arena.GetResult()
	arena.buffer.offset += size
	return ptr, nil
}

// GetRemainder returns current offset modulo alignment.
func (arena *MemoryArena[T]) GetRemainder(alignment uintptr) int {
	return arena.buffer.offset % int(alignment)
}

// alignOffset moves the offset forward to satisfy alignment requirements.
func (arena *MemoryArena[T]) alignOffset(alignment uintptr) {
	if rem := arena.GetRemainder(alignment); rem != 0 {
		arena.buffer.offset += int(alignment - uintptr(rem))
	}
}

// nextOffset returns what the offset would be after allocating size bytes.
func (arena *MemoryArena[T]) nextOffset(size int) int {
	return arena.buffer.offset + size
}

// Free zeroes out the buffer.
func (arena *MemoryArena[T]) Free() {
	for i := range arena.buffer.memory {
		arena.buffer.memory[i] = 0
	}
}

// Reset clears the buffer and resets offset to 0.
func (arena *MemoryArena[T]) Reset() {
	arena.Free()
	arena.buffer.offset = 0
}

// AllocateNewValue allocates space and copies obj into the arena.
// Returns a pointer to the allocated memory or an error.
func (arena *MemoryArena[T]) AllocateNewValue(size int, obj T) (unsafe.Pointer, error) {
	ptr, err := arena.Allocate(size)
	if err != nil {
		return nil, err
	}
	*(*T)(ptr) = obj
	return ptr, nil
}

// AllocateObject allocates memory for obj of type T, copying its value into the arena.
func (arena *MemoryArena[T]) AllocateObject(obj interface{}) (unsafe.Pointer, error) {
	value, ok := obj.(T)
	if !ok {
		return nil, ErrInvalidType
	}
	size := int(unsafe.Sizeof(value))
	ptr, err := arena.AllocateNewValue(size, value)
	if err != nil {
		return nil, err
	}
	return ptr, nil
}

// Resize resets the arena to newSize, discarding all data. Returns ErrInvalidSize if newSize <= 0.
func (arena *MemoryArena[T]) Resize(newSize int) error {
	if newSize <= 0 {
		return ErrInvalidSize
	}
	arena.Free()
	arena.buffer.memory = make([]byte, newSize)
	arena.buffer.size = newSize
	arena.buffer.offset = 0
	return nil
}

// ResizePreserve resizes the arena to newSize, preserving existing data. Errors if newSize < used.
func (arena *MemoryArena[T]) ResizePreserve(newSize int) error {
	if newSize <= 0 {
		return ErrInvalidSize
	}
	used := arena.buffer.offset
	if used > newSize {
		return ErrNewSizeTooSmall
	}
	arena.SetNewMemory(newSize, used)
	return nil
}

// SetNewMemory replaces the buffer with a new one of newSize, copying used bytes.
func (arena *MemoryArena[T]) SetNewMemory(newSize int, used int) {
	newMem := make([]byte, newSize)
	copy(newMem, arena.buffer.memory[:used])
	arena.buffer.memory = newMem
	arena.buffer.size = newSize
}
