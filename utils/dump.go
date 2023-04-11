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

	// the vector is the normal of the split plane.
	//   Thus, the "T norm" is extracted from the children array address.??
	return "TODO: Fix normal of the split plane."
}
