package interfaces

import (
	"io"
	"unsafe"
)

type IndexAllocator interface {
	Open(fqFile string) (AllocatedIndex, error)
	Get(fqFile string) (AllocatedIndex, bool)
}

type AllocatedIndex interface {
	io.Closer
	Ptr() unsafe.Pointer
	Size() int64
}

// BuildIndexAllocator is an allocator whilst building an index.
type BuildIndexAllocator interface {
	// Free frees the memory allocated by the allocator.
	Free()
	//Reallocate will allocate/reallocate memory to fit the given size.
	Reallocate(byteSize int) unsafe.Pointer
}
