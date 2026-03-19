//go:build linux || darwin || freebsd || openbsd || netbsd

package memory

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// OnDiskBuildAllocator implements BuildIndexAllocator by writing directly
// to a memory-mapped file. The file grows as the index is built, avoiding
// the need for a separate Save step.
type OnDiskBuildAllocator struct {
	file *os.File
	data []byte
	ptr  unsafe.Pointer
	size int
}

// NewOnDiskBuildAllocator creates a new allocator that builds directly
// into the specified file. The file is created or truncated.
func NewOnDiskBuildAllocator(fileName string) (*OnDiskBuildAllocator, error) {
	_ = os.Remove(fileName)

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("on-disk build: create file: %w", err)
	}

	return &OnDiskBuildAllocator{file: file}, nil
}

func (a *OnDiskBuildAllocator) Reallocate(byteSize int) unsafe.Pointer {
	if a.data != nil && byteSize <= a.size {
		return a.ptr
	}

	// Unmap old region if present.
	if a.data != nil {
		_ = syscall.Munmap(a.data)
		a.data = nil
		a.ptr = nil
	}

	// Grow the file.
	if err := a.file.Truncate(int64(byteSize)); err != nil {
		panic(fmt.Sprintf("on-disk build: truncate to %d: %v", byteSize, err))
	}

	// Memory-map the file read/write.
	data, err := syscall.Mmap(
		int(a.file.Fd()),
		0,
		byteSize,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		panic(fmt.Sprintf("on-disk build: mmap %d bytes: %v", byteSize, err))
	}

	a.data = data
	a.ptr = unsafe.Pointer(&data[0])
	a.size = byteSize

	return a.ptr
}

// Truncate shrinks the backing file to the given size after the build is
// complete, removing excess space from the growth factor.
func (a *OnDiskBuildAllocator) Truncate(size int64) error {
	// Must unmap before truncating.
	if a.data != nil {
		_ = syscall.Munmap(a.data)
		a.data = nil
		a.ptr = nil
	}
	if a.file == nil {
		return nil
	}
	return a.file.Truncate(size)
}

func (a *OnDiskBuildAllocator) Free() {
	if a.data != nil {
		_ = syscall.Munmap(a.data)
		a.data = nil
		a.ptr = nil
	}
	if a.file != nil {
		_ = a.file.Close()
		a.file = nil
	}
	a.size = 0
}

// FileName returns the path of the backing file, or empty string if closed.
func (a *OnDiskBuildAllocator) FileName() string {
	if a.file == nil {
		return ""
	}
	return a.file.Name()
}
