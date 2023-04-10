package memory

import (
	"arena"
	"unsafe"
)

type BuildIndexMemoryArenaAllocatorImpl struct {
	currentArena *arena.Arena
	ptr          unsafe.Pointer
	ptrSize      int
	is64         bool
	active       []byte
}

func NewBuildIndexMemoryArenaAllocator() *BuildIndexMemoryArenaAllocatorImpl {
	return &BuildIndexMemoryArenaAllocatorImpl{
		currentArena: arena.NewArena(),
		is64:         unsafe.Sizeof(int(0)) == 8,
	}
}

func (a *BuildIndexMemoryArenaAllocatorImpl) Free() {
	if a.currentArena != nil {
		a.currentArena.Free()
	}

	a.currentArena = nil
	a.ptr = nil
	a.ptrSize = 0
	a.active = nil
}

func (a *BuildIndexMemoryArenaAllocatorImpl) Reallocate(byteSize int) unsafe.Pointer {
	// Create a new arena to do the "reallocate" to
	ar := arena.NewArena()

	data := arena.MakeSlice[byte](ar, byteSize, byteSize)
	ptr := unsafe.Pointer(unsafe.SliceData(data))

	if a.ptrSize > 0 {
		// Copy the memory from old arena to new arena
		if a.is64 {
			copy((*[1 << 49]byte)(ptr)[:a.ptrSize], (*[1 << 49]byte)(a.ptr)[:a.ptrSize])
		} else {
			copy((*[1 << 31]byte)(ptr)[:a.ptrSize], (*[1 << 31]byte)(a.ptr)[:a.ptrSize])
		}
	}

	if a.currentArena != nil {
		a.currentArena.Free()
	}

	a.currentArena = ar
	a.ptr = ptr
	a.ptrSize = byteSize
	a.active = data

	return ptr
}
