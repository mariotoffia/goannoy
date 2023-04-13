package memory

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
)

type mmapIndexAllocator struct {
	indexes map[string]*mmapIndexAllocation
}

type mmapIndexAllocation struct {
	fqFile string
	file   *os.File
	ptr    unsafe.Pointer
	size   int64
	data   []byte
	parent *mmapIndexAllocator
}

func (mi *mmapIndexAllocation) Ptr() unsafe.Pointer {
	return mi.ptr
}

func (mi *mmapIndexAllocation) Size() int64 {
	return mi.size
}

// Implements `io.Closer` interface
func (mi *mmapIndexAllocation) Close() error {
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

func MmapIndexAllocator() *mmapIndexAllocator {
	return &mmapIndexAllocator{
		indexes: map[string]*mmapIndexAllocation{},
	}
}

func (mm *mmapIndexAllocator) Get(fqFile string) (interfaces.IndexMemory, bool) {
	index, ok := mm.indexes[fqFile]
	return index, ok
}

func (mm *mmapIndexAllocator) Open(fqFile string) (interfaces.IndexMemory, error) {
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

	mi := &mmapIndexAllocation{
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
