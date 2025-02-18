package memoryArena

import (
	"fmt"
	"reflect"
	"unsafe"
)

// interface for MemoryArena and ConcurrentArena behaviours
type Arena interface {
	Reset()
	AllocateObject(obj interface{}) (unsafe.Pointer, error)
}

// Object is being allocated in the Arena.
func AllocateObject[T any](arena Arena, obj T) (unsafe.Pointer, error) {
	return arena.AllocateObject(obj)
}

// Resetting the Arena.
func Reset(arena Arena) {
	arena.Reset()
}

// NewObject allocate memory through AllocateObject, returns pointer to T or error handle.
func NewObject[T any](arena Arena, obj T) (*T, error) {
	ptr, err := AllocateObject(arena, obj)
	if err != nil {
		return nil, err
	}
	return (*T)(ptr), nil
}

// AppendSlice appends object to slice and returns pointer to slice or error handle.
func AppendSlice[T any](obj *T, arena Arena, slice *[]T) (*[]T, error) {
	*slice = append(*slice, *obj)
	ptr, err := AllocateObject(arena, slice)
	if err != nil {
		return nil, err
	}
	return (*[]T)(ptr), nil

}

// InsertMap inserts object to hashmap and returns pointer to hashmap or error handle.
func InsertMap[T any](obj *T, arena Arena, hashmap *map[string]T, key string) (*map[string]T, error) {
	(*hashmap)[key] = *obj
	ptr, err := AllocateObject(arena, hashmap)
	if err != nil {
		return nil, err
	}
	return (*map[string]T)(ptr), nil

}

// SetNewValue sets new value to pointer.
func SetNewValue(ptr *unsafe.Pointer, obj interface{}) (unsafe.Pointer, error) {
	if ptr == nil || *ptr == nil {
		return nil, fmt.Errorf("invalid pointer provided")
	}
	newValue := reflect.NewAt(
		reflect.TypeOf(obj),
		*ptr,
	).Elem()
	newValue.Set(reflect.ValueOf(obj))
	return *ptr, nil
}
