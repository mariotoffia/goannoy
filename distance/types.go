package distance

type VectorType interface {
	float32 | float64
}

type DistancePreprocessor[TV VectorType] interface {
	// PreProcess will pre-process the data before it is used for distance calculations.
	PreProcess(nodes []Node[TV], vectorLength int)
}

type NodeFactory[TV VectorType] interface {
	NewNode(vectorLength int) Node[TV]
}
