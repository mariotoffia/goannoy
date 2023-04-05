package distance

import (
	"unsafe"

	"github.com/mariotoffia/goannoy/vector"
)

type Direction int

const (
	Left  Direction = 0
	Right Direction = 1
)

type Node[TV VectorType] interface {
	// GetRawVector returns the raw vector that you have to know the length in order
	// to safely access it.
	//
	// ```go
	// func ShowVector[TV VectorType](node *NodeImpl[TV]) {
	// vectorLength := 1536
	// ptr := unsafe.Pointer(node.GetFirstFloat32Ptr())
	// size := unsafe.Sizeof(TV(0))
	//
	//	for i := uintptr(0); i < uintptr(vectorLength); i++ {
	//	 f := *(*TV)(unsafe.Pointer(ptr + i*size))
	//	 fmt.Printf("n.v[%d] = %.2f\n", i, f)
	//	}
	//
	// ```
	GetRawVector() *TV
	// GetVector will allocate a slice header and point to the raw vector.
	//
	// CAUTION: This is a wasteful and should be used with care since it will
	// allocate a new slice header every time!
	//
	// It uses the _vectorLength_ to know how many elements to set as length in
	// the slice. Be *careful* to use the correct length, otherwise it may corrupt
	// the memory upon writes in the vector.
	GetVector(vectorLength int) []TV
	// SetVector will set the vector to the given slice. It does this by copying
	// the slice contents to the raw vector. It uses the _vectorLength_ to
	// check that the length of the slice is correct. If not, it will `panic`.
	SetVector(v []TV, vectorLength int)
	// GetChildren returns all children both on lef and right side.
	GetChildren() [2][]int32
	// SetChildren will set the children on the given direction. The children are
	// indexes to the nodes.
	SetChildren(dir Direction, children []int32)
	GetNumberOfDescendants() int32
	SetNumberOfDescendants(n int32)
	// Normalize will normalize the vector
	Normalize(vectorLength int)
	// CopyNodeTo will copy this Node contents to dst Node
	CopyNodeTo(dst Node[TV], vectorLength int)
	// InitNode will initialize the node. Depending on the implementation
	// it will do different things.
	InitNode(vectorLength int)
	// Distance calculates the distance from this to the _to_ `Node`.
	Distance(to Node[TV], vectorLength int) TV
	IsDataPoint() bool
	// Size returns the size of the implementation in bytes.
	Size(vectorLength int) int
	// MaxDescendants is the amount of children this node can have.
	MaxDescendants() int
}

// NodeImpl base type for all nodes
type NodeImpl[TV VectorType] struct {
	nDescendants int32
	children     [2][]int32
	v            [0]TV
}

// NewNodeImpl creates a new NodeImpl from memory _mem_.
func NewNodeImpl[TV VectorType](mem *byte) *NodeImpl[TV] {
	n := (*NodeImpl[TV])(unsafe.Pointer(mem))
	return n
}

func (n *NodeImpl[TV]) GetSize(vectorLength int) int {
	return int(
		unsafe.Sizeof(n.nDescendants) +
			(uintptr(len(n.children)) * unsafe.Sizeof(n.children[0])) +
			(uintptr(len(n.v)) * uintptr(vectorLength)),
	)
}

func (n *NodeImpl[TV]) GetRawVector() *TV {
	return (*TV)(unsafe.Pointer(&n.v))
}

func (n *NodeImpl[TV]) GetVector(vectorLength int) []TV {
	return unsafe.Slice((*TV)(unsafe.Pointer(&n.v)), vectorLength)
}

func (n *NodeImpl[TV]) SetVector(v []TV, vectorLength int) {
	if len(v) != vectorLength {
		panic("Vector length mismatch")
	}

	dst := unsafe.Pointer(&n.v)
	src := unsafe.Pointer(&v[0])

	// Copy data from the input slice to the underlying memory
	size := uintptr(len(v)) * unsafe.Sizeof(TV(0))

	// 1 << 30 is really fooling the compiler to think it is 1GB memory.
	// The purpose of this trick is to create a slice that shares the
	// underlying memory with the input pointers without knowing the
	// exact size of the memory region at compile time.
	copy((*[1 << 30]byte)(dst)[:size], (*[1 << 30]byte)(src)[:size])
}

func (n *NodeImpl[TV]) GetChildren() [2][]int32 {
	return n.children
}

func (n *NodeImpl[TV]) SetChildren(dir Direction, children []int32) {
	n.children[int(dir)] = children
}

func (n *NodeImpl[TV]) GetNumberOfDescendants() int32 {
	return n.nDescendants
}

func (n *NodeImpl[TV]) SetNumberOfDescendants(nDescendants int32) {
	n.nDescendants = nDescendants
}

func (n *NodeImpl[TV]) IsDataPoint() bool {
	return n.nDescendants == 1
}

func (n *NodeImpl[TV]) Normalize(vectorLength int) {
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
