package memoryArena

import (
	"fmt"
	"reflect"
	"sync"
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

func Reset(arena Arena) {
	arena.Reset()
}

// NewObject allocate memory through AllocateObject, returns pointer to T or error handle.
func NewObject[T any](arena Arena, obj T) (*T, error) {
	mutex := sync.Mutex{}
	mutex.Lock()
	ptr, err := AllocateObject(arena, obj)
	if err != nil {
		return nil, err
	}
	defer mutex.Unlock()
	return (*T)(ptr), nil
}

func AppendSlice[T any](obj *T, arena Arena, slice *[]T) (*[]T, error) {
	*slice = append(*slice, *obj)
	ptr, err := AllocateObject(arena, slice)
	if err != nil {
		return nil, err
	}
	return (*[]T)(ptr), nil

}

func SetNewValue(ptr *unsafe.Pointer, obj interface{}) (unsafe.Pointer, error) {
	newValue := reflect.NewAt(
		reflect.TypeOf(obj),
		*ptr,
	).Elem()
	if &newValue == nil {
		return nil, fmt.Errorf("Cannot set new value")
	}
	newValue.Set(reflect.ValueOf(obj))
	return *ptr, nil
}
