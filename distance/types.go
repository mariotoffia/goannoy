package distance

type VectorType interface {
	float32 | float64
}

type Distance[TV VectorType] interface {
	Distance(x, y Node[TV]) TV
	// CopyNode will copy the Node src to dst
	CopyNode(dst, src Node[TV])
	// Normalize will normalize the vector
	Normalize(node Node[TV])
	// InitNode will initialize the node by setting the norm to
	// the value based on the distance type.
	InitNode(node Node[TV])
}
