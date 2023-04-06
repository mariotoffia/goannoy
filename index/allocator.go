package index

import "unsafe"

type Allocator interface {
	// Free frees the memory allocated by the allocator.
	Free()
	Reallocate(byteSize int) unsafe.Pointer
}
