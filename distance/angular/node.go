package angular

import (
	"math"
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/vector"
)

// AngularNodeImpl is the node implementation for the angular distance.
//
// Note from the author:
//
// We store a binary tree where each node has two things
// - A vector associated with it
// - Two children
// All nodes occupy the same amount of memory
type AngularNodeImpl[TV interfaces.VectorType] struct {
	// n_descendants is the number of descendants of this node.
	//
	// * All nodes with n_descendants == 1 are leaf nodes.
	// * For nodes with n_descendants == 1 the vector is a data point.
	// * For nodes with n_descendants > K the vector is the normal of the split plane.
	//   Thus, the "T norm" is extracted from the children array address.
	n_descendants int
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
	children [2]int
	v        [0]TV
}

func (n *AngularNodeImpl[TV]) Size(vectorLength int) int {
	// _s = offsetof(Node, v) + _f * sizeof(T); // Size of each node
	return int(unsafe.Offsetof(n.v) + (uintptr(vectorLength) * unsafe.Sizeof(TV(0))))
}

func (n *AngularNodeImpl[TV]) GetRawVector() *TV {
	return (*TV)(unsafe.Pointer(&n.v))
}

func (n *AngularNodeImpl[TV]) GetVector(vectorLength int) []TV {
	return unsafe.Slice((*TV)(unsafe.Pointer(&n.v)), vectorLength)
}

func (n *AngularNodeImpl[TV]) SetVector(v []TV) {
	dst := unsafe.Pointer(&n.v)
	src := unsafe.Pointer(&v[0])
	size := uintptr(len(v)) * unsafe.Sizeof(TV(0))

	copy((*[1 << 31]byte)(dst)[:size], (*[1 << 31]byte)(src)[:size])
}

func (n *AngularNodeImpl[TV]) GetRawChildren() *int {
	return (*int)(unsafe.Pointer(&n.children))
}

func (n *AngularNodeImpl[TV]) GetChildren() []int {
	if n.n_descendants == 0 {
		return interfaces.EmptyChildren
	}

	return unsafe.Slice((*int)(unsafe.Pointer(&n.children)), n.n_descendants)
}

func (n *AngularNodeImpl[TV]) SetChildren(children []int) {
	dst := unsafe.Pointer(&n.children)
	src := unsafe.Pointer(&children[0])
	size := uintptr(len(children)) * unsafe.Sizeof(int(0))

	copy((*[1 << 31]byte)(dst)[:size], (*[1 << 31]byte)(src)[:size])
}

func (n *AngularNodeImpl[TV]) GetNumberOfDescendants() int {
	return n.n_descendants
}

func (n *AngularNodeImpl[TV]) SetNumberOfDescendants(nDescendants int) {
	n.n_descendants = nDescendants
}

func (n *AngularNodeImpl[TV]) IsDataPoint() bool {
	return n.n_descendants == 1
}

// HasIndexes returns true if the n.children is being used to store indexes.
func (n *AngularNodeImpl[TV]) HasIndexes(vectorLength int) bool {
	return n.n_descendants > 1 && n.n_descendants < n.MaxNumChildren(vectorLength)
}

func (n *AngularNodeImpl[TV]) Normalize(vectorLength int) {
	raw := n.GetRawVector()
	norm := TV(vector.GetNormUnsafe(raw, vectorLength))

	if norm > 0 {
		ptr := unsafe.Pointer(raw)
		size := int(unsafe.Sizeof(TV(0)))

		for i := 0; i < vectorLength; i++ {
			f := (*TV)(unsafe.Pointer(unsafe.Add(ptr, i*size)))
			*f /= norm
		}
	}
}

// InitNode will initialize the node by setting the norm to
// the value based on the distance type.
func (n *AngularNodeImpl[TV]) InitNode(vectorLength int) {
	// write norm to the children array address
	norm := (*TV)(unsafe.Pointer(&n.children))
	*norm = vector.DotUnsafe(n.GetRawVector(), n.GetRawVector(), vectorLength)
}

func (n *AngularNodeImpl[TV]) CopyNodeTo(dst interfaces.Node[TV], vectorLength int) {
	dstPtr := (unsafe.Pointer(dst.(*AngularNodeImpl[TV])))
	srcPtr := unsafe.Pointer(n)
	size := n.Size(vectorLength)

	// Copy data from the input slice to the underlying memory
	copy((*[1 << 31]byte)(dstPtr)[:size], (*[1 << 31]byte)(srcPtr)[:size])
}

func (x *AngularNodeImpl[TV]) MaxNumChildren(vectorLength int) int {
	// _K = (S) (((size_t) (_s - offsetof(Node, children))) / sizeof(S));
	return int((uintptr(x.Size(vectorLength)) - unsafe.Offsetof(x.children)) / unsafe.Sizeof(x.children[0]))
}

func (x *AngularNodeImpl[TV]) Distance(to interfaces.Node[TV], vectorLength int) TV {
	t := to.(*AngularNodeImpl[TV])
	pp := *(*TV)(unsafe.Pointer(&x.children))
	qq := *(*TV)(unsafe.Pointer(&t.children))

	if pp == 0 {
		pp = vector.DotUnsafe(x.GetRawVector(), x.GetRawVector(), vectorLength)
	}

	if qq == 0 {
		qq = vector.DotUnsafe(t.GetRawVector(), t.GetRawVector(), vectorLength)
	}

	pq := vector.DotUnsafe(x.GetRawVector(), t.GetRawVector(), vectorLength)
	ppqq := pp * qq

	if ppqq > 0 {
		return 2.0 - 2.0*pq/TV(math.Sqrt(float64(ppqq)))
	}
	return 2.0
}
