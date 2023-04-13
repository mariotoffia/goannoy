package memory

import (
	"fmt"
	"io"
	"os"
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
)

type fileIndexerAllocator struct {
	indexes map[string]*fileIndexAllocation
}

type fileIndexAllocation struct {
	fqFile string
	ptr    unsafe.Pointer
	data   []byte
	size   int64
	parent *fileIndexerAllocator
}

func (fi *fileIndexAllocation) Ptr() unsafe.Pointer {
	return fi.ptr
}

func (fi *fileIndexAllocation) Size() int64 {
	return fi.size
}

// Implements `io.Closer` interface
func (fi *fileIndexAllocation) Close() error {
	delete(fi.parent.indexes, fi.fqFile)

	fi.data = nil
	fi.ptr = nil

	return nil
}

func FileIndexMemoryAllocator() *fileIndexerAllocator {
	return &fileIndexerAllocator{
		indexes: map[string]*fileIndexAllocation{},
	}
}

func (mm *fileIndexerAllocator) Get(fqFile string) (interfaces.IndexMemory, bool) {
	index, ok := mm.indexes[fqFile]
	return index, ok
}

func (mm *fileIndexerAllocator) Open(fqFile string) (interfaces.IndexMemory, error) {
	file, err := os.Open(fqFile)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	data := make([]byte, fileInfo.Size())

	var bytesRead int64

	for bytesRead < fileInfo.Size() {
		n, err := io.ReadFull(file, data[bytesRead:])
		if err != nil && err != io.ErrUnexpectedEOF {
			panic(fmt.Sprintf("Failed to read file: %v", err))
		}
		bytesRead += int64(n)
	}

	if err != nil {
		file.Close()
		return nil, err
	}

	fi := &fileIndexAllocation{
		parent: mm,
		fqFile: fqFile,
		size:   fileInfo.Size(),
		ptr:    unsafe.Pointer(&data[0]),
		data:   data,
	}

	mm.indexes[fqFile] = fi
	return fi, nil
}
