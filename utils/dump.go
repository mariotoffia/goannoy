package utils

import (
	"fmt"

	"github.com/mariotoffia/goannoy/interfaces"
)

func DumpNode[TV interfaces.VectorType](node interfaces.Node[TV], vectorLength int) string {
	descendants := node.GetNumberOfDescendants()

	if descendants == 1 {
		// Leaf node
		vec := node.GetVector(vectorLength)
		return fmt.Sprintf("LeafNode: %v", vec)
	}

	if descendants <= node.MaxNumChildren(vectorLength) {
		// Internal node
		children := node.GetChildren()
		return fmt.Sprintf("InternalNode - children: %v", children)
	}

	// the vector is the normal of the split plane.
	//   Thus, the "T norm" is extracted from the children array address.
	//TODO:
	return "TODO: FGix normal of the split plane."
}
