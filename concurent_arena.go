package memoryArena

import (
	"sync"
	"unsafe"
)

type ConcurrentArena struct {
	mutex sync.Mutex
	arena MemoryArena
}

func NewConcurrentArena(arena MemoryArena) *ConcurrentArena {
	return &ConcurrentArena{
		arena: arena,
	}
}

func (arena *ConcurrentArena) Allocate(size int) unsafe.Pointer {
	arena.mutex.Lock()
	ptr := arena.arena.Allocate(size)
	arena.mutex.Unlock()
	return ptr
}

func (arena *ConcurrentArena) Reset() {
	arena.mutex.Lock()
	arena.arena.Reset()
	arena.mutex.Unlock()
}
