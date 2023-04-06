package index

import (
	"arena"
	"unsafe"
)

type ArenaAllocatorImpl struct {
	currentArena *arena.Arena
	ptr          unsafe.Pointer
	ptrSize      int
}

func NewArenaAllocator() *ArenaAllocatorImpl {
	return &ArenaAllocatorImpl{
		currentArena: arena.NewArena(),
	}
}

func (a *ArenaAllocatorImpl) Free() {
	if a.currentArena != nil {
		a.currentArena.Free()
	}

	a.currentArena = nil
}

func (a *ArenaAllocatorImpl) Reallocate(byteSize int) unsafe.Pointer {
	// Create a new arena to do the "reallocate" to
	ar := arena.NewArena()

	data := arena.MakeSlice[byte](ar, byteSize, byteSize)
	ptr := unsafe.Pointer(&data[0])

	if a.ptrSize > 0 {
		// Copy the memory from old arena to new arena
		copy((*[1 << 30]byte)(ptr)[:a.ptrSize], (*[1 << 30]byte)(a.ptr)[:a.ptrSize])
	}

	if a.currentArena != nil {
		a.currentArena.Free()
	}

	a.currentArena = ar
	a.ptr = ptr
	a.ptrSize = byteSize

	return ptr
}
