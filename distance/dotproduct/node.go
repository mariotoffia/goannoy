package dotproduct

import (
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
)

// DotProductNodeImpl is the node implementation for the dot product distance.
//
// This is an extension of the Angular node with an extra attribute for the scaled norm.
type DotProductNodeImpl[TV interfaces.VectorType] struct {
	n_descendants interfaces.ItemID
	children      [2]interfaces.ItemID
	dot_factor    TV
	norm          TV
	built         bool
	v             [0]TV
}

func (n *DotProductNodeImpl[TV]) GetRawVector() *TV {
	return (*TV)(unsafe.Pointer(&n.v))
}

func (n *DotProductNodeImpl[TV]) GetVector(vectorLength int) []TV {
	return unsafe.Slice((*TV)(unsafe.Pointer(&n.v)), vectorLength)
}

func (n *DotProductNodeImpl[TV]) SetVector(v []TV) {
	dst := unsafe.Pointer(&n.v)
	src := unsafe.Pointer(unsafe.SliceData(v))
	size := uintptr(len(v)) * unsafe.Sizeof(TV(0))

	copy((*[1 << 30]byte)(dst)[:size], (*[1 << 30]byte)(src)[:size])
}

func (n *DotProductNodeImpl[TV]) GetRawChildren() *interfaces.ItemID {
	return (*interfaces.ItemID)(unsafe.Pointer(&n.children))
}

func (n *DotProductNodeImpl[TV]) GetChildren() []interfaces.ItemID {
	if n.n_descendants == 0 {
		return nil
	}

	return unsafe.Slice((*interfaces.ItemID)(unsafe.Pointer(&n.children)), n.n_descendants)
}

func (n *DotProductNodeImpl[TV]) SetChildren(children []interfaces.ItemID) {
	dst := unsafe.Pointer(&n.children)
	src := unsafe.Pointer(unsafe.SliceData(children))
	size := uintptr(len(children)) * unsafe.Sizeof(n.children[0])

	copy((*[1 << 30]byte)(dst)[:size], (*[1 << 30]byte)(src)[:size])
}

func (n *DotProductNodeImpl[TV]) GetNumberOfDescendants() interfaces.ItemID {
	return n.n_descendants
}

func (n *DotProductNodeImpl[TV]) SetNumberOfDescendants(nDescendants interfaces.ItemID) {
	n.n_descendants = nDescendants
}

func (n *DotProductNodeImpl[TV]) GetNorm() TV {
	return n.norm
}

func (n *DotProductNodeImpl[TV]) SetNorm(norm TV) {
	n.norm = norm
}
