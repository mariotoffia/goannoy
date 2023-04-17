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

type Distance[TV VectorType, TIX IndexTypes] interface {
	// PreProcess will pre-process the data before it is used for distance calculations.
	//
	// The _nodes_ is a pointer to the beginning of the memory where the nodes are stored.
	PreProcess(nodes unsafe.Pointer, node_count TIX)
	// Distance calculates the distance from _x_ to _y_ `Node`.
	Distance(x Node[TV, TIX], y Node[TV, TIX]) TV
	// Normalize will normalize the vector for the _node_.
	Normalize(node Node[TV, TIX])
	// Margin will return the margin for the node.
	Margin(n Node[TV, TIX], y []TV) TV
	// CreateSplit will write to split node _m_ based on the _children_ nodes. The _nodeSize_ is the
	// size of the memory a `Node[TV,TIX]` will occupy. The _vectorLength_ is the length of the vector
	// the node will hold.
	CreateSplit(
		children []Node[TV, TIX],
		nodeSize TIX,
		random Random[TIX],
		m Node[TV, TIX],
	)
	// Side determines which side of the children indices to use when a split is made.
	Side(
		m Node[TV, TIX],
		v []TV,
		random Random[TIX],
	) Side
	// MapNodeToMemory will map the node to existing memory and use that for storage.
	MapNodeToMemory(mem unsafe.Pointer, itemIndex TIX) Node[TV, TIX]
	PQDistance(distance, margin TV, side Side) TV
	PQInitialValue() TV
	// InitNode will initialize the node. Depending on the implementation
	// it will do different things.
	InitNode(node Node[TV, TIX])
	// MaxNumChildren is the max number of descendants to fit into node by overwriting
	// the vector space.
	MaxNumChildren() TIX
	// NodeSize is the size of the allocated memory for the node. Each node occupy the same
	// amount of memory.
	NodeSize() TIX
	// VectorLength is the length of the vector the this distance operates on.
	VectorLength() TIX
}
