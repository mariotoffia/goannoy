package distance

type Node[TV VectorType] interface {
	GetVector() []TV
	SetVector(v []TV)
	GetNumberOfDescendants() int32
	SetNumberOfDescendants(n int32)
	GetNorm() TV
	SetNorm(n TV)
}

// NodeImpl base type for all nodes
type NodeImpl[TV VectorType] struct {
	nDescendants int32
	norm         TV
	v            []TV
}

func (n *NodeImpl[TV]) GetVector() []TV {
	return n.v
}

func (n *NodeImpl[TV]) SetVector(v []TV) {
	n.v = v
}

func (n *NodeImpl[TV]) GetNumberOfDescendants() int32 {
	return n.nDescendants
}

func (n *NodeImpl[TV]) SetNumberOfDescendants(nDescendants int32) {
	n.nDescendants = nDescendants
}

func (n *NodeImpl[TV]) GetNorm() TV {
	return n.norm
}

func (n *NodeImpl[TV]) SetNorm(norm TV) {
	n.norm = norm
}
