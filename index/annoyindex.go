package index

import (
	"fmt"

	"github.com/mariotoffia/goannoy/distance"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/random"
	"github.com/mariotoffia/goannoy/vector"
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
	TIdx int,
	TV distance.VectorType,
	TRandType random.RandomTypes,
	TRand random.Random[TRandType]] struct {
	vectorLength int
	_s           int
	// _n_items is how many nodes exists in the index.
	_n_items     TIdx
	_nodes       []distance.Node[TV]
	_n_nodes     TIdx
	_nodes_size  TIdx
	_roots       []TIdx
	_K           TIdx
	_seed        random.Random[TRandType]
	indexLoaded  bool
	_verbose     bool
	_fd          int
	_on_disk     bool
	_built       bool
	nodeFactory  distance.NodeFactory[TV]
	preprocessor distance.DistancePreprocessor[TV]
	buildPolicy  interfaces.AnnoyIndexBuildPolicy
}

func NewAnnoyIndexImpl[
	TIdx int,
	TV distance.VectorType,
	TRandType random.RandomTypes,
	TRand random.Random[TRandType]](
	vectorLength int,
	random random.Random[TRandType],
	nodeFactory distance.NodeFactory[TV],
	preprocessor distance.DistancePreprocessor[TV],
	buildPolicy interfaces.AnnoyIndexBuildPolicy,
) *AnnoyIndexImpl[TIdx, TV, TRandType, TRand] {
	//
	index := &AnnoyIndexImpl[TIdx, TV, TRandType, TRand]{
		vectorLength: vectorLength,
		_seed:        random,
		_s:           0, // TODO: I think we should skip this to not use unsafe...
		_verbose:     false,
		_built:       false,
		// TODO: (TIdx) (((size_t) (_s - offsetof(Node, children))) / sizeof(TIdx));
		// TODO: Max number of descendants to fit into node
		_K:          0,
		nodeFactory: nodeFactory,
		buildPolicy: buildPolicy,
	}

	index.reinitialize()

	return index
}

// AddItem adds an item to the index. The ownership of the vector _v_ is taken
// by this function.
func (idx *AnnoyIndexImpl[S, TV, TRandType, TRand]) AddItem(
	index S,
	v [vector.ANNOYLIB_V_ARRAY_SIZE]TV,
) {
	if idx.indexLoaded {
		panic("Can't add items to a loaded index")
	}

	node := idx.nodeFactory.NewNode(idx.vectorLength)
	node.SetNumberOfDescendants(1)
	node.SetVector(v)
	node.InitNode(idx.vectorLength)
}

func (idx *AnnoyIndexImpl[TIdx, TV, TRandType, TRand]) Build(treesPerThread int) {
	if idx.indexLoaded {
		panic("Can't build a loaded index")
	}

	if idx._built {
		panic("Index already built")
	}

	idx.preprocessor.PreProcess(idx._nodes, idx.vectorLength)
	idx._n_nodes = idx._n_items

	idx.buildPolicy.Build(idx, treesPerThread, treesPerThread)
}

func (idx *AnnoyIndexImpl[TIdx, TV, TRandType, TRand]) ThreadBuild(
	treesPerThread, threadIdx int,
	threadedBuildPolicy interfaces.AnnoyIndexBuildPolicy) {
	// Each thread needs its own seed, otherwise each thread would be building the same tree(s)
	rnd := idx._seed.CloneAndReset()

	var threadRoots []TIdx

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

		if idx._verbose {
			fmt.Printf("pass %d...\n", len(threadRoots))
		}

		var indices []TIdx
		threadedBuildPolicy.LockSharedNodes()
		for i := TIdx(0); i < idx._n_items; i++ {
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

func (idx *AnnoyIndexImpl[TIdx, TV, TRandType, TRand]) makeTree(
	indices []TIdx, isRoot bool,
	rnd random.Random[TRandType],
	threadedBuildPolicy interfaces.AnnoyIndexBuildPolicy,
) TIdx {

	if len(indices) == 1 && !isRoot {
		return indices[0]
	}

	if TIdx(len(indices)) <= idx._K && (!isRoot || idx._n_items <= idx._K || len(indices) == 1) {
		threadedBuildPolicy.LockNNodes()
		item := idx._n_nodes
		idx._n_nodes++
		threadedBuildPolicy.UnlockNNodes()

		threadedBuildPolicy.LockSharedNodes()
		m := idx._nodes[item]
		if isRoot {
			m.SetNumberOfDescendants(int32(idx._n_items))
		} else {
			m.SetNumberOfDescendants(int32(len(indices)))
		}

		if len(indices) > 0 {
			children := [vector.NUM_CHILDREN]int32{}
			for i, index := range indices {
				children[i] = int32(index)
			}

			m.SetChildren(children)
		}

		threadedBuildPolicy.UnlockSharedNodes()
		return item
	}

	// Add logic for handling split nodes, which is missing in the provided code snippet.
	return -1
}

func (idx *AnnoyIndexImpl[S, TV, TRandType, TRand]) reinitialize() {
	idx._fd = 0
	idx._nodes = []distance.Node[TV]{}
	idx.indexLoaded = false
	idx._n_items = 0
	idx._n_nodes = 0
	idx._nodes_size = 0
	idx._on_disk = false
	idx._seed = idx._seed.CloneAndReset()
	idx._roots = []S{}
}
