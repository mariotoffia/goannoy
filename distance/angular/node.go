package angular

import (
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
)

// AngularNodeImpl is the node implementation for the angular distance.
//
// Note from the author:
//
// We store a binary tree where each node has two things
// - A vector associated with it
// - Two children
// All nodes occupy the same amount of memory
type AngularNodeImpl[TV interfaces.VectorType, TIX interfaces.IndexTypes] struct {
	// n_descendants is the number of descendants of this node.
	//
	// * All nodes with n_descendants == 1 are leaf nodes.
	// * For nodes with n_descendants == 1 the vector is a data point.
	// * For nodes with n_descendants > K the vector is the normal of the split plane.
	//   Thus, the "T norm" is extracted from the children array address.
	n_descendants TIX
	// children will contain indexes to other nodes when n_descendants > 1.
	//
	// A memory optimization is when n_descendants >= 2 (and less than K, where K is the
	// calculated maximum number of descendants that can fit instead of the vector).
	// In that case no vector is stored and the memory is used for children only instead.
	//
	// The original C++ was an union. In this implementation I use the children to store norm
	// when n_descendants == 1.
	// ```cpp
	// union {
	//	S children[2];
	//	T norm;
	// };
	// ```
	children [2]TIX
	v        [0]TV
}

func (n *AngularNodeImpl[TV, TIX]) GetRawVector() *TV {
	return (*TV)(unsafe.Pointer(&n.v))
}

func (n *AngularNodeImpl[TV, TIX]) GetVector(vectorLength TIX) []TV {
	return unsafe.Slice((*TV)(unsafe.Pointer(&n.v)), vectorLength)
}

func (n *AngularNodeImpl[TV, TIX]) SetVector(v []TV) {
	dst := unsafe.Pointer(&n.v)
	src := unsafe.Pointer(unsafe.SliceData(v))
	size := uintptr(len(v)) * unsafe.Sizeof(TV(0))

	copy((*[1 << 30]byte)(dst)[:size], (*[1 << 30]byte)(src)[:size])
}

func (n *AngularNodeImpl[TV, TIX]) GetRawChildren() *TIX {
	return (*TIX)(unsafe.Pointer(&n.children))
}

func (n *AngularNodeImpl[TV, TIX]) GetChildren() []TIX {
	if n.n_descendants == 0 {
		return nil
	}

	return unsafe.Slice((*TIX)(unsafe.Pointer(&n.children)), n.n_descendants)
}

func (n *AngularNodeImpl[TV, TIX]) SetChildren(children []TIX) {
	dst := unsafe.Pointer(&n.children)
	src := unsafe.Pointer(unsafe.SliceData(children))
	size := uintptr(len(children)) * unsafe.Sizeof(n.children[0])

	copy((*[1 << 30]byte)(dst)[:size], (*[1 << 30]byte)(src)[:size])
}

func (n *AngularNodeImpl[TV, TIX]) GetNumberOfDescendants() TIX {
	return n.n_descendants
}

func (n *AngularNodeImpl[TV, TIX]) SetNumberOfDescendants(nDescendants TIX) {
	n.n_descendants = nDescendants
}

func (n *AngularNodeImpl[TV, TIX]) GetNorm() TV {
	return *(*TV)(unsafe.Pointer(&n.children))
}

func (n *AngularNodeImpl[TV, TIX]) SetNorm(norm TV) {
	*(*TV)(unsafe.Pointer(&n.children)) = norm
}
