package index

import (
	"github.com/mariotoffia/goannoy/distance"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/random"
)

// AnnoyIndexImpl is the actual index for all vectors.
//
// A Note from the author:
//
// We use random projection to build a forest of binary trees of all items.
// Basically just split the hyperspace into two sides by a hyperplane,
// then recursively split each of those subtrees etc.
// We create a tree like this q times. The default q is determined automatically
// in such a way that we at most use 2x as much memory as the vectors take.

type AnnoyIndexImpl[
	TV distance.VectorType,
	TRandType random.RandomTypes,
	TRand random.Random[TRandType]] struct {
	vectorLength int
	nodeSize     int
	// _n_items is how many nodes exists in the index.
	_n_items       int
	_nodes         []distance.Node[TV]
	_n_nodes       int
	_nodes_size    int
	_roots         []int
	maxDescendants int
	random         random.Random[TRandType]
	indexLoaded    bool
	_on_disk       bool
	indexBuilt     bool
	nodeFactory    distance.NodeFactory[TV]
	preprocessor   distance.DistancePreprocessor[TV]
	buildPolicy    interfaces.AnnoyIndexBuildPolicy
}

func NewAnnoyIndexImpl[
	TV distance.VectorType,
	TRandType random.RandomTypes,
	TRand random.Random[TRandType]](
	vectorLength int,
	random random.Random[TRandType],
	nodeFactory distance.NodeFactory[TV],
	preprocessor distance.DistancePreprocessor[TV],
	buildPolicy interfaces.AnnoyIndexBuildPolicy,
) *AnnoyIndexImpl[TV, TRandType, TRand] {
	// Create a single node to query it for sizes
	node := nodeFactory.NewNode(vectorLength)

	index := &AnnoyIndexImpl[TV, TRandType, TRand]{
		vectorLength:   vectorLength,                      // _f
		random:         random,                            // _seed
		nodeSize:       node.Size(vectorLength),           // _s
		maxDescendants: node.MaxNumChildren(vectorLength), // _K
		indexBuilt:     false,                             // _built
		nodeFactory:    nodeFactory,
		buildPolicy:    buildPolicy,
	}

	index.reinitialize()

	return index
}

// VectorLength returns the vector length of the index.
func (idx *AnnoyIndexImpl[TV, TRandType, TRand]) VectorLength() int {
	return idx.vectorLength
}

// AddItem adds an item to the index. The ownership of the vector _v_ is taken
// by this function.
func (idx *AnnoyIndexImpl[TV, TRandType, TRand]) AddItem(index int, v []TV) {
	if idx.indexLoaded {
		panic("Can't add items to a loaded index")
	}

	node := idx.nodeFactory.NewNode(idx.vectorLength)
	node.SetNumberOfDescendants(1)
	node.SetVector(v)
	node.InitNode(idx.vectorLength)

	idx._nodes = append(idx._nodes, node)

	if index >= idx._n_items {
		idx._n_items = index + 1
	}
}

func (idx *AnnoyIndexImpl[TV, TRandType, TRand]) Build(treesPerThread int) {
	if idx.indexLoaded {
		panic("Can't build a loaded index")
	}

	if idx.indexBuilt {
		panic("Index already built")
	}

	idx.preprocessor.PreProcess(idx._nodes, idx._n_items, idx.vectorLength)
	idx._n_nodes = idx._n_items

	idx.buildPolicy.Build(idx, treesPerThread, treesPerThread)
}

func (idx *AnnoyIndexImpl[TV, TRandType, TRand]) ThreadBuild(
	treesPerThread, threadIdx int,
	threadedBuildPolicy interfaces.AnnoyIndexBuildPolicy,
) {
	rnd := idx.random.CloneAndReset()

	// Each thread needs its own seed, otherwise each thread would be building the same tree(s)
	rnd.SetSeed(rnd.GetSeed() + TRandType(threadIdx))

	var threadRoots []int

	for {
		if treesPerThread == -1 {
			threadedBuildPolicy.LockNNodes()
			if idx._n_nodes >= 2*idx._n_items {
				threadedBuildPolicy.UnlockNNodes()
				break
			}
			threadedBuildPolicy.UnlockNNodes()
		} else {
			if len(threadRoots) >= treesPerThread {
				break
			}
		}

		var indices []int
		threadedBuildPolicy.LockSharedNodes()
		for i := 0; i < idx._n_items; i++ {
			node := idx._nodes[i]

			if node.GetNumberOfDescendants() >= 1 {
				indices = append(indices, i)
			}
		}

		threadedBuildPolicy.UnlockSharedNodes()

		threadRoots = append(threadRoots, idx.makeTree(indices, true, rnd, threadedBuildPolicy))
	}

	threadedBuildPolicy.LockRoots()
	idx._roots = append(idx._roots, threadRoots...)
	threadedBuildPolicy.UnlockRoots()
}

func (idx *AnnoyIndexImpl[TV, TRandType, TRand]) makeTree(
	indices []int, isRoot bool,
	rnd random.Random[TRandType],
	threadedBuildPolicy interfaces.AnnoyIndexBuildPolicy,
) int {

	if len(indices) == 1 && !isRoot {
		return indices[0]
	}

	if len(indices) <= idx.maxDescendants && (!isRoot || idx._n_items <= idx.maxDescendants || len(indices) == 1) {
		threadedBuildPolicy.LockNNodes()
		item := idx._n_nodes
		idx._n_nodes++
		threadedBuildPolicy.UnlockNNodes()

		threadedBuildPolicy.LockSharedNodes()
		m := idx._nodes[item]
		if isRoot {
			m.SetNumberOfDescendants(idx._n_items)
		} else {
			m.SetNumberOfDescendants(len(indices))
		}

		if len(indices) > 0 {
			children := make([]int, len(indices))
			copy(children, indices)

			m.SetChildren(children)
		}

		threadedBuildPolicy.UnlockSharedNodes()
		return item
	}

	// TODO: Add logic for handling split nodes, which is missing in the provided code snippet.
	return -1
}

func (idx *AnnoyIndexImpl[TV, TRandType, TRand]) reinitialize() {
	idx._nodes = nil
	idx.indexLoaded = false
	idx._n_items = 0
	idx._n_nodes = 0
	idx._nodes_size = 0
	idx._on_disk = false
	idx.random = idx.random.CloneAndReset()
	idx._roots = nil
}
