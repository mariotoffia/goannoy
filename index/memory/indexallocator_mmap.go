package memory

import (
	"os"
	"syscall"
	"unsafe"
)

type Mmap struct {
	indexes map[string]*mappedIndex
}

type mappedIndex struct {
	fqFile string
	file   *os.File
	ptr    uintptr
	size   int64
	data   []byte
	parent *Mmap
}

func (mi *mappedIndex) Ptr() uintptr {
	return mi.ptr
}

// Implements `io.Closer` interface
func (mi *mappedIndex) Close() error {
	delete(mi.parent.indexes, mi.fqFile)

	var err error

	if mi.data != nil {
		err = syscall.Munmap(mi.data)
	}

	err2 := mi.file.Close()

	if err == nil {
		err = err2
	}

	return err
}

func NewMmap() *Mmap {
	return &Mmap{
		indexes: map[string]*mappedIndex{},
	}
}

func (mm *Mmap) Get(fqFile string) (IndexMemory, bool) {
	index, ok := mm.indexes[fqFile]
	return index, ok
}

func (mm *Mmap) Open(fqFile string) (IndexMemory, error) {
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
		ptr:    uintptr(unsafe.Pointer(&data[0])),
		data:   data,
	}

	mm.indexes[fqFile] = mi
	return mi, nil
}
