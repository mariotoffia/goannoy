package dotproduct

import (
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
)

// DotProductNodeImpl is the node implementation for the dot product distance.
//
// This is an extension of the Angular node with an extra attribute for the scaled norm.
type DotProductNodeImpl[TV interfaces.VectorType, TIX interfaces.IndexTypes] struct {
	n_descendants TIX
	children      [2]TIX
	dot_factor    TV
	v             [0]TV
}

func (n *DotProductNodeImpl[TV, TIX]) GetRawVector() *TV {
	return (*TV)(unsafe.Pointer(&n.v))
}

func (n *DotProductNodeImpl[TV, TIX]) GetVector(vectorLength TIX) []TV {
	return unsafe.Slice((*TV)(unsafe.Pointer(&n.v)), vectorLength)
}

func (n *DotProductNodeImpl[TV, TIX]) SetVector(v []TV) {
	dst := unsafe.Pointer(&n.v)
	src := unsafe.Pointer(unsafe.SliceData(v))
	size := uintptr(len(v)) * unsafe.Sizeof(TV(0))

	copy((*[1 << 30]byte)(dst)[:size], (*[1 << 30]byte)(src)[:size])
}

func (n *DotProductNodeImpl[TV, TIX]) GetRawChildren() *TIX {
	return (*TIX)(unsafe.Pointer(&n.children))
}

func (n *DotProductNodeImpl[TV, TIX]) GetChildren() []TIX {
	if n.n_descendants == 0 {
		return nil
	}

	return unsafe.Slice((*TIX)(unsafe.Pointer(&n.children)), n.n_descendants)
}

func (n *DotProductNodeImpl[TV, TIX]) SetChildren(children []TIX) {
	dst := unsafe.Pointer(&n.children)
	src := unsafe.Pointer(unsafe.SliceData(children))
	size := uintptr(len(children)) * unsafe.Sizeof(n.children[0])

	copy((*[1 << 30]byte)(dst)[:size], (*[1 << 30]byte)(src)[:size])
}

func (n *DotProductNodeImpl[TV, TIX]) GetNumberOfDescendants() TIX {
	return n.n_descendants
}

func (n *DotProductNodeImpl[TV, TIX]) SetNumberOfDescendants(nDescendants TIX) {
	n.n_descendants = nDescendants
}

func (n *DotProductNodeImpl[TV, TIX]) GetNorm() TV {
	return *(*TV)(unsafe.Pointer(&n.children))
}

func (n *DotProductNodeImpl[TV, TIX]) SetNorm(norm TV) {
	*(*TV)(unsafe.Pointer(&n.children)) = norm
}
