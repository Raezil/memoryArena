package memoryArena

import (
	"unsafe"
)

// Arena defines the common behavior for memory arenas.
type Arena interface {
	Reset()
	AllocateObject(obj interface{}) (unsafe.Pointer, error)
}

// AllocateObject allocates an object in the Arena.
func AllocateObject[T any](arena Arena, obj T) (unsafe.Pointer, error) {
	return arena.AllocateObject(obj)
}

// Reset resets the Arena.
func Reset(arena Arena) {
	arena.Reset()
}

// NewObject allocates memory for an object, returning a pointer or an error.
func NewObject[T any](arena Arena, obj T) (*T, error) {
	ptr, err := AllocateObject(arena, obj)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}

// AppendSlice appends an object to a slice and allocates space for the updated slice.
func AppendSlice[T any](obj *T, arena Arena, slice *[]T) (*[]T, error) {
	*slice = append(*slice, *obj)
	// Pass the dereferenced slice so that AllocateObject sees []T (matching the arena’s type)
	ptr, err := AllocateObject(arena, *slice)
	if err != nil {
		return nil, err
	}
	return (*[]T)(ptr), nil
}

// InsertMap inserts an object into a map under a given key and allocates space for the updated map.
func InsertMap[T any](obj *T, arena Arena, hashmap *map[string]T, key string) (*map[string]T, error) {
	(*hashmap)[key] = *obj
	// Pass the dereferenced map so that AllocateObject sees map[string]T
	ptr, err := AllocateObject(arena, *hashmap)
	if err != nil {
		return nil, err
	}
	return (*map[string]T)(ptr), nil
}
