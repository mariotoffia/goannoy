package interfaces

var EmptyChildren = []int{}

type Node[TV VectorType] interface {
	// GetRawVector returns the raw vector that you have to know the length in order
	// to safely access it.
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
	// the slice contents to the raw vector.
	SetVector(v []TV)
	// GetRawChildren returns the raw children that you have to know the length in order
	// to safely access it.
	GetRawChildren() *int
	// GetChildren returns all children indexes (if n_descendants > 1 && n_descendants <= K).
	// This will allocate a new slice header and point to the raw children.
	GetChildren() []int
	// SetChildren will copy the children slice to the node.
	SetChildren(children []int)
	GetNumberOfDescendants() int
	SetNumberOfDescendants(n int)
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
	// Size returns the size of the node implementation in bytes.
	Size(vectorLength int) int
	// MaxNumChildren is the max number of descendants to fit into node by overwriting
	// the vector space.
	MaxNumChildren(vectorLength int) int
}
