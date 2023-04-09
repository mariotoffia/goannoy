package memory

import (
	"fmt"
	"io"
	"os"
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
)

type fileIndexer struct {
	indexes map[string]*fileIndex
}

type fileIndex struct {
	fqFile string
	ptr    unsafe.Pointer
	data   []byte
	size   int64
	parent *fileIndexer
}

func (fi *fileIndex) Ptr() unsafe.Pointer {
	return fi.ptr
}

func (fi *fileIndex) Size() int64 {
	return fi.size
}

// Implements `io.Closer` interface
func (fi *fileIndex) Close() error {
	delete(fi.parent.indexes, fi.fqFile)

	fi.data = nil
	fi.ptr = nil

	return nil
}

func FileIndexMemoryAllocator() *fileIndexer {
	return &fileIndexer{
		indexes: map[string]*fileIndex{},
	}
}

func (mm *fileIndexer) Get(fqFile string) (interfaces.IndexMemory, bool) {
	index, ok := mm.indexes[fqFile]
	return index, ok
}

func (mm *fileIndexer) Open(fqFile string) (interfaces.IndexMemory, error) {
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

	fi := &fileIndex{
		parent: mm,
		fqFile: fqFile,
		size:   fileInfo.Size(),
		ptr:    unsafe.Pointer(&data[0]),
		data:   data,
	}

	mm.indexes[fqFile] = fi
	return fi, nil
}
