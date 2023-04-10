package interfaces

import "unsafe"

type Allocator interface {
	// Free frees the memory allocated by the allocator.
	Free()
	//Reallocate will allocate/reallocate memory to fit the given size.
	Reallocate(byteSize int) unsafe.Pointer
}
