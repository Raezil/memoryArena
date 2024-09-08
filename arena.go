package memoryArena

import "unsafe"

type Arena interface {
	Allocate(size int) unsafe.Pointer
	Reset()
	AllocateObject(obj interface{}) (unsafe.Pointer, error)
}

func Allocate(arena Arena, size int) unsafe.Pointer {
	return arena.Allocate(size)
}

func AllocateObject(arena Arena, obj interface{}) (unsafe.Pointer, error) {
	return arena.AllocateObject(obj)
}

func Reset(arena Arena) {
	arena.Reset()
}
