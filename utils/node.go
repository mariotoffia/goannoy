package utils

import (
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
)

func GetPtr[TV interfaces.VectorType](node interfaces.Node[TV]) unsafe.Pointer {
	// iface is a fake interface type used to access the underlying data of an interface.
	type iface struct {
		typ, data unsafe.Pointer
	}

	nodeInterface := (*iface)(unsafe.Pointer(&node))
	return nodeInterface.data
}

// CopyNode copies the source node to the destination node. Note that the destination node
// must be of the same type and take up the same amount of memory as the source node.
func CopyNode[TV interfaces.VectorType](dst, src interfaces.Node[TV], size int) {
	ptrSrc := GetPtr(src)
	ptrDst := GetPtr(dst)

	// Copy the memory from the source to the destination.
	copy((*[1 << 30]byte)(ptrDst)[:size], (*[1 << 30]byte)(ptrSrc)[:size])
}
