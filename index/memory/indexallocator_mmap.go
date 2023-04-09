package memory

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
)

type mmap struct {
	indexes map[string]*mappedIndex
}

type mappedIndex struct {
	fqFile string
	file   *os.File
	ptr    unsafe.Pointer
	size   int64
	data   []byte
	parent *mmap
}

func (mi *mappedIndex) Ptr() unsafe.Pointer {
	return mi.ptr
}

func (mi *mappedIndex) Size() int64 {
	return mi.size
}

// Implements `io.Closer` interface
func (mi *mappedIndex) Close() error {
	delete(mi.parent.indexes, mi.fqFile)

	var err error

	if mi.data != nil {
		err = syscall.Munmap(mi.data)
	}

	if mi.file != nil {
		if err2 := mi.file.Close(); err == nil && err2 != nil {
			err = err2
		}
	}

	mi.data = nil
	mi.ptr = nil
	mi.file = nil

	return err
}

func MmapIndexAllocator() *mmap {
	return &mmap{
		indexes: map[string]*mappedIndex{},
	}
}

func (mm *mmap) Get(fqFile string) (interfaces.IndexMemory, bool) {
	index, ok := mm.indexes[fqFile]
	return index, ok
}

func (mm *mmap) Open(fqFile string) (interfaces.IndexMemory, error) {
	file, err := os.Open(fqFile)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	size := stat.Size()

	data, err := syscall.Mmap(
		int(file.Fd()),
		0,
		int(size),
		syscall.PROT_READ,
		syscall.MAP_SHARED,
	)

	if err != nil {
		file.Close()
		return nil, err
	}

	mi := &mappedIndex{
		parent: mm,
		fqFile: fqFile,
		file:   file,
		size:   size,
		ptr:    unsafe.Pointer(&data[0]),
		data:   data,
	}

	mm.indexes[fqFile] = mi
	return mi, nil
}
