package memory

import "io"

type IndexMemoryAllocator interface {
	Open(fqFile string) (IndexMemory, error)
	Get(fqFile string) (IndexMemory, bool)
}

type IndexMemory interface {
	io.Closer
	Ptr() uintptr
}
