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
	PreProcess(nodes unsafe.Pointer, node_count, vectorLength int)
	// NormalizeDistance will normalize the distance to a value between 0 and 1.
	NormalizedDistance(distance TV) TV
	// CreateSplit will write to split node _m_ based on the _children_ nodes. The _nodeSize_ is the
	// size of the memory a `Node[TV]` will occupy. The _vectorLength_ is the length of the vector
	// the node will hold.
	CreateSplit(
		children []Node[TV],
		vectorLength, nodeSize int,
		random Random[TR],
		m Node[TV],
	)
	// Side determines which side of the children indices to use when a split is made.
	Side(
		m Node[TV],
		v []TV,
		random Random[TR],
		vectorLength int,
	) Side
	// MapNodeToMemory will map the node to existing memory and use that for storage.
	MapNodeToMemory(mem unsafe.Pointer, itemIndex, vectorLength int) Node[TV]
	// NewNodeFromGC will create a new node from the garbage collector managed memory.
	NewNodeFromGC(vectorLength int) Node[TV]
}
