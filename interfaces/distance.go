package interfaces

import (
	"unsafe"
)

type Side int

const (
	SideLeft  Side = 0
	SideRight Side = 1
)

type VectorType interface {
	float32 | float64
}

type Distance[TV VectorType, TR RandomTypes] interface {
	// PreProcess will pre-process the data before it is used for distance calculations.
	//
	// The _nodes_ is a pointer to the beginning of the memory where the nodes are stored.
	PreProcess(nodes unsafe.Pointer, node_count int)
	// Distance calculates the distance from _x_ to _y_ `Node`.
	Distance(x Node[TV], y Node[TV]) TV
	// Normalize will normalize the vector for the _node_.
	Normalize(node Node[TV])
	// Margin will return the margin for the node.
	Margin(n Node[TV], y []TV) TV
	// CreateSplit will write to split node _m_ based on the _children_ nodes. The _nodeSize_ is the
	// size of the memory a `Node[TV]` will occupy. The _vectorLength_ is the length of the vector
	// the node will hold.
	CreateSplit(
		children []Node[TV],
		nodeSize int,
		random Random[TR],
		m Node[TV],
	)
	// Side determines which side of the children indices to use when a split is made.
	Side(
		m Node[TV],
		v []TV,
		random Random[TR],
	) Side
	// MapNodeToMemory will map the node to existing memory and use that for storage.
	MapNodeToMemory(mem unsafe.Pointer, itemIndex int) Node[TV]
	PQDistance(distance, margin TV, side Side) TV
	PQInitialValue() TV
	// InitNode will initialize the node. Depending on the implementation
	// it will do different things.
	InitNode(node Node[TV])
	// MaxNumChildren is the max number of descendants to fit into node by overwriting
	// the vector space.
	MaxNumChildren() int
	// NodeSize is the size of the allocated memory for the node. Each node occupy the same
	// amount of memory.
	NodeSize() int
	// VectorLength is the length of the vector the node will hold.
	VectorLength() int
}
