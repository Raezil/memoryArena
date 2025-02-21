package memoryArena

// NewObject allocates an object in the arena and returns a pointer.
func NewObject[T any](arena *MemoryArena[T], obj T) (*T, error) {
	return arena.AllocateNewValue(obj)
}

// AppendSlice appends an element to a slice and allocates the new slice in the arena.
func AppendSlice[T any](arena *MemoryArena[[]T], element T, slice []T) (*[]T, error) {
	newSlice := append(slice, element)
	return arena.AllocateNewValue(newSlice)
}

// InsertMap inserts a key-value pair into a map and allocates the updated map in the arena.
func InsertMap[T any](arena *MemoryArena[map[string]T], key string, value T, m map[string]T) (*map[string]T, error) {
	if m == nil {
		m = make(map[string]T)
	}
	m[key] = value
	return arena.AllocateNewValue(m)
}
