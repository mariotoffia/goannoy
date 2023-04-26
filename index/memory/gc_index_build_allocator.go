package memory

import (
	"unsafe"
)

type GoGCIndexBuildAllocatorImpl struct {
	ptr     unsafe.Pointer
	ptrSize int
	is64    bool
	memory  []byte
}

func GoGCIndexAllocator() *GoGCIndexBuildAllocatorImpl {
	return &GoGCIndexBuildAllocatorImpl{
		is64: unsafe.Sizeof(int(0)) == 8,
	}
}

func (a *GoGCIndexBuildAllocatorImpl) Free() {
	a.ptr = nil
	a.ptrSize = 0
	a.memory = nil
}

func (a *GoGCIndexBuildAllocatorImpl) Reallocate(byteSize int) unsafe.Pointer {
	if a.memory != nil && byteSize < a.ptrSize {
		// No new memory needed
		return a.ptr
	}

	data := make([]byte, byteSize)
	ptr := unsafe.Pointer(unsafe.SliceData(data))

	if a.ptrSize > 0 {
		// Copy the memory from old arena to new arena
		if a.is64 {
			copy((*[1 << 49]byte)(ptr)[:a.ptrSize], (*[1 << 49]byte)(a.ptr)[:a.ptrSize])
		} else {
			copy((*[1 << 31]byte)(ptr)[:a.ptrSize], (*[1 << 31]byte)(a.ptr)[:a.ptrSize])
		}
	}

	a.ptr = ptr
	a.ptrSize = byteSize
	a.memory = data

	return ptr
}
