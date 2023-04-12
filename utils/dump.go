package utils

import (
	"fmt"

	"github.com/mariotoffia/goannoy/interfaces"
)

func DumpNode[TV interfaces.VectorType, TR interfaces.RandomTypes](
	distance interfaces.Distance[TV, TR],
	node interfaces.Node[TV],
) string {
	descendants := node.GetNumberOfDescendants()

	if descendants == 1 {
		// Leaf node
		vec := node.GetVector(distance.VectorLength())
		return fmt.Sprintf("LeafNode: %v", vec)
	}

	if descendants <= distance.MaxNumChildren() {
		// Internal node
		children := node.GetChildren()
		return fmt.Sprintf("InternalNode - children: %v", children)
	}

	norm := node.GetNorm()
	vec := node.GetVector(distance.VectorLength())

	return fmt.Sprintf("SplitNode - norm: %v, vector: %v", norm, vec)
}
