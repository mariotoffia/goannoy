package memory

import (
	"fmt"
	"io"
	"os"
	"unsafe"
)

type FileIndexer struct {
	indexes map[string]*fileIndex
}

type fileIndex struct {
	fqFile string
	ptr    uintptr
	data   []byte
	parent *FileIndexer
}

func (mi *fileIndex) Ptr() uintptr {
	return mi.ptr
}

// Implements `io.Closer` interface
func (mi *fileIndex) Close() error {
	delete(mi.parent.indexes, mi.fqFile)
	mi.data = nil
	return nil
}

func NewFileIndexer() *FileIndexer {
	return &FileIndexer{
		indexes: map[string]*fileIndex{},
	}
}

func (mm *FileIndexer) Get(fqFile string) (IndexMemory, bool) {
	index, ok := mm.indexes[fqFile]
	return index, ok
}

func (mm *FileIndexer) Open(fqFile string) (IndexMemory, error) {
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
		ptr:    uintptr(unsafe.Pointer(&data[0])),
		data:   data,
	}

	mm.indexes[fqFile] = fi
	return fi, nil
}
