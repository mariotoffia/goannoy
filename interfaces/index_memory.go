package interfaces

import (
	"io"
	"unsafe"
)

type IndexMemoryAllocator interface {
	Open(fqFile string) (IndexMemory, error)
	Get(fqFile string) (IndexMemory, bool)
}

type IndexMemory interface {
	io.Closer
	Ptr() unsafe.Pointer
	Size() int64
}
