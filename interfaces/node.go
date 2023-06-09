package interfaces

var EmptyChildren = []int{}

type Node[TV VectorType, TIX IndexTypes] interface {
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
	GetVector(vectorLength TIX) []TV
	// SetVector will set the vector to the given slice. It does this by copying
	// the slice contents to the raw vector.
	SetVector(v []TV)
	// GetRawChildren returns the raw children that you have to know the length in order
	// to safely access it.
	GetRawChildren() *TIX
	// GetChildren returns all children indexes (if n_descendants > 1 && n_descendants <= K).
	// This will allocate a new slice header and point to the raw children.
	GetChildren() []TIX
	// SetChildren will copy the children slice to the node.
	SetChildren(children []TIX)
	GetNumberOfDescendants() TIX
	SetNumberOfDescendants(n TIX)
	GetNorm() TV
	SetNorm(norm TV)
}
