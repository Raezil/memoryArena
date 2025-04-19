package memoryArena

import "errors"

var (
	ErrOutOfMemory     = errors.New("memory arena: out of memory")
	ErrArenaFull       = errors.New("memory arena: insufficient space")
	ErrInvalidSize     = errors.New("memory arena: size must be greater than 0")
	ErrNewSizeTooSmall = errors.New("memory arena: new size is smaller than current usage")
	ErrInvalidType     = errors.New("memory arena: invalid object type for this arena")
)
