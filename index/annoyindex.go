package index

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/utils"
)

const reallocation_factor = float64(1.5)

// AnnoyIndexImpl is the actual index for all vectors.
//
// A Note from the authors https://github.com/spotify/annoy
//
// We use random projection to build a forest of binary trees of all items.
// Basically just split the hyperspace into two sides by a hyperplane,
// then recursively split each of those subtrees etc.
// We create a tree like this q times. The default q is determined automatically
// in such a way that we at most use 2x as much memory as the vectors take.
type AnnoyIndexImpl[
	TV interfaces.VectorType,
	TIX interfaces.IndexTypes] struct {
	vectorLength TIX
	// nodeSize the the complete size of the node in bytes.
	nodeSize TIX
	// _n_items is how many nodes exists in the index.
	_n_items TIX
	_nodes   unsafe.Pointer
	_n_nodes TIX
	// _nodes_size is the number of nodes that has been allocated.
	// Total size is _node_size * nodeSize
	_nodes_size TIX
	_roots      []TIX
	// batchMaxNNS is the maximum of indexes that a query can possibly create.
	// This is updated each time a index is loaded.
	batchMaxNNS          int
	logVerbose           bool
	maxDescendants       TIX
	random               interfaces.Random[TIX]
	indexLoaded          bool
	indexBuilt           bool
	distance             interfaces.Distance[TV, TIX]
	buildPolicy          interfaces.AnnoyIndexBuildPolicy
	allocator            interfaces.Allocator
	indexMemoryAllocator interfaces.IndexMemoryAllocator
	indexMemory          interfaces.IndexMemory
}

// New create a new index instance based on the _TV_ for the vector
// and _TIX_ for the index type. When done use the `io.Closer.Close()` to clean up
// any resources.
//
// The _vectorLength_ is the number of elements in the vector that this index handles.
// The _random_ is the random generator to use for the index. The _distance_ is the
// distance functions to use for the index (_see sub-packages under distance/ for different
// types_). The _buildPolicy_ is the policy to use when building the index. Those are located
// in the `policy` package. The _allocator_ is the allocator to use while building the index.
// Allocators reside in `package memory`.
//
// NOTE: It is possible to provide with a positive integer for _hintNumIndexes_ to pre-allocate
// to speed up the index creation.
//
// The _indexMemoryAllocator_ is the allocator to use for the index memory when loading it from
// file. See `package memory` for more information. It is possible to output to stdout by setting
// _logVerbose_ to `true`. This will output the progress of the index creation.
//
// Use `AddIndex` and when done, `Build` to build the index. `Save` the index, and thus is then
// ready to be used for lookups.
func New[
	TV interfaces.VectorType,
	TIX interfaces.IndexTypes](
	random interfaces.Random[TIX],
	distance interfaces.Distance[TV, TIX],
	buildPolicy interfaces.AnnoyIndexBuildPolicy,
	allocator interfaces.Allocator,
	indexMemoryAllocator interfaces.IndexMemoryAllocator,
	logVerbose bool,
	hintNumIndexes TIX,
) interfaces.AnnoyIndex[TV, TIX] {
	//
	index := &AnnoyIndexImpl[TV, TIX]{
		vectorLength:         distance.VectorLength(),   // _f
		random:               random,                    // _seed
		nodeSize:             distance.NodeSize(),       // _s
		maxDescendants:       distance.MaxNumChildren(), // _K
		indexBuilt:           false,                     // _built
		logVerbose:           logVerbose,                // _verbose
		distance:             distance,
		allocator:            allocator,
		buildPolicy:          buildPolicy,
		indexMemoryAllocator: indexMemoryAllocator,
	}

	// Pre-allocate memory for the index if hintNumIndexes is set > 0
	if hintNumIndexes > 0 {
		allocator.Reallocate(int(float64(distance.NodeSize()*hintNumIndexes) * reallocation_factor))
	}

	return index
}

// Implements `io.Closer` interface
func (idx *AnnoyIndexImpl[TV, TIX]) Close() error {
	var err error

	if idx.indexMemory != nil {
		err = idx.indexMemory.Close()
		idx.indexMemory = nil
	}

	if idx.allocator != nil {
		idx.allocator.Free()
	}

	idx._nodes = nil
	idx.indexLoaded = false
	idx._n_items = 0
	idx._n_nodes = 0
	idx._nodes_size = 0
	idx.random = idx.random.CloneAndReset()
	idx._roots = nil

	return err
}

// VectorLength returns the vector length of the index.
func (idx *AnnoyIndexImpl[TV, TIX]) VectorLength() TIX {
	return idx.vectorLength
}

func (idx *AnnoyIndexImpl[TV, TIX]) GetItem(itemIndex TIX) []TV {
	return idx.getNode(itemIndex).GetVector(idx.vectorLength)
}

func (idx *AnnoyIndexImpl[TV, TIX]) AddItem(itemIndex TIX, v []TV) {
	if idx.indexLoaded {
		panic("Can't add items to a loaded index")
	}

	if idx.vectorLength != TIX(len(v)) {
		panic(fmt.Sprintf("Vector length mismatch: %d != %d", idx.vectorLength, len(v)))
	}

	// Ensure that we have enough memory for the new node
	idx.allocateSize(itemIndex+1, nil)

	// Map the node onto the memory
	node := idx.getNode(itemIndex)

	// Initialize the node with the vector
	node.SetNumberOfDescendants(1)
	node.SetVector(v)
	idx.distance.InitNode(node)

	// Is new spot?
	if itemIndex >= idx._n_items {
		idx._n_items = itemIndex + 1
	}

	if idx.logVerbose {
		fmt.Printf(
			"added itemIndex:%d node - %s\n", itemIndex, utils.DumpNode(idx.distance, node),
		)
	}
}

func (idx *AnnoyIndexImpl[TV, TIX]) Build(numberOfTrees, numWorkers int) {
	if idx.indexLoaded {
		panic("Can't build a loaded index")
	}

	if idx.indexBuilt {
		panic("Index already built")
	}

	// Give the preprocessor a chance to process the nodes before building the index
	idx.distance.PreProcess(idx._nodes, idx._n_items)

	idx._n_nodes = idx._n_items

	idx.buildPolicy.Build(idx, numberOfTrees, numWorkers)

	// Also, copy the roots into the last segment of the array
	// This way we can load them faster without reading the whole file
	idx.allocateSize(idx._n_nodes+TIX(len(idx._roots)), nil)

	for i := TIX(0); i < TIX(len(idx._roots)); i++ {
		dst := idx.getNode(idx._n_nodes + i)
		src := idx.getNode(idx._roots[i])

		utils.CopyNode(dst, src, idx.vectorLength)

		if idx.logVerbose {
			fmt.Printf(
				"added roots[i=%d]:%d node - %s\n", i, idx._roots[i], utils.DumpNode(idx.distance, dst),
			)
		}
	}

	idx._n_nodes += TIX(len(idx._roots))
	idx.indexBuilt = true

	idx.batchMaxNNS = -1

	for i := TIX(0); i < idx._n_nodes; i++ {
		nd := idx.getNode(i)

		nDescendants := nd.GetNumberOfDescendants()

		if nDescendants == 1 && i < idx._n_items {
			idx.batchMaxNNS++
		} else if nDescendants <= idx.maxDescendants {
			idx.batchMaxNNS += len(nd.GetChildren())
		}
	}

	if idx.logVerbose {
		fmt.Println("Max NNS:", idx.batchMaxNNS)
	}
}

// ThreadBuild is called from the build policy to build the index.
func (idx *AnnoyIndexImpl[TV, TIX]) ThreadBuild(
	treesPerWorker, workerIdx int,
	threadedBuildPolicy interfaces.AnnoyIndexBuildPolicy,
) {
	rnd := idx.random.CloneAndReset()

	// Each worker needs its own seed, otherwise each worker would be building the same tree(s)
	rnd.SetSeed(rnd.GetSeed() + TIX(workerIdx))

	var threadRoots []TIX

	for {
		if treesPerWorker == -1 {
			threadedBuildPolicy.LockNNodes()
			if idx._n_nodes >= 2*idx._n_items {
				threadedBuildPolicy.UnlockNNodes()
				break
			}
			threadedBuildPolicy.UnlockNNodes()
		} else {
			if len(threadRoots) >= treesPerWorker {
				break
			}
		}

		var indices []TIX

		threadedBuildPolicy.LockSharedNodes()

		for i := TIX(0); i < idx._n_items; i++ {
			node := idx.getNode(i)

			if node.GetNumberOfDescendants() >= 1 {
				indices = append(indices, i)
			}
		}

		threadedBuildPolicy.UnlockSharedNodes()

		threadRoots = append(
			threadRoots,
			idx.makeTree(indices, true, rnd, threadedBuildPolicy),
		)
	}

	threadedBuildPolicy.LockRoots()
	idx._roots = append(idx._roots, threadRoots...)
	threadedBuildPolicy.UnlockRoots()
}

func (idx *AnnoyIndexImpl[TV, TIX]) getNode(index TIX) interfaces.Node[TV, TIX] {
	return idx.distance.MapNodeToMemory(idx._nodes, index)
}

func (idx *AnnoyIndexImpl[TV, TIX]) makeTree(
	indices []TIX, isRoot bool,
	rnd interfaces.Random[TIX],
	threadedBuildPolicy interfaces.AnnoyIndexBuildPolicy,
) TIX {
	// The basic rule is that if we have <= maxDescendants items, then it's a leaf node, otherwise it's a split node.
	// There's some regrettable complications caused by the problem that root nodes have to be "special":
	// 1. We identify root nodes by the arguable logic that _n_items == n->n_descendants,
	//    regardless of how many descendants they actually have
	//
	// 2. Root nodes with only 1 child need to be a "dummy" parent
	//
	// 3. Due to the _n_items "hack", we need to be careful with the cases where _n_items <= _K or _n_items > _K
	if len(indices) == 1 && !isRoot {
		return indices[0]
	}

	lenIdx := TIX(len(indices))
	if lenIdx <= idx.maxDescendants &&
		(!isRoot || idx._n_items <= idx.maxDescendants || lenIdx == 1) {
		// Ensure we have memory for the new node
		threadedBuildPolicy.LockNNodes()
		idx.allocateSize(idx._n_nodes+1, threadedBuildPolicy)

		item := idx._n_nodes
		idx._n_nodes++
		threadedBuildPolicy.UnlockNNodes()

		threadedBuildPolicy.LockSharedNodes()

		m := idx.getNode(item)

		if isRoot {
			m.SetNumberOfDescendants(idx._n_items)
		} else {
			m.SetNumberOfDescendants(lenIdx)
		}

		if len(indices) > 0 {
			children := make([]TIX, len(indices))
			copy(children, indices)

			m.SetChildren(children)
		}

		threadedBuildPolicy.UnlockSharedNodes()

		if idx.logVerbose {
			fmt.Printf("added 1:node[item=%d] - %s\n", item, utils.DumpNode(idx.distance, m))
		}

		return item
	}

	threadedBuildPolicy.LockSharedNodes()

	var children []interfaces.Node[TV, TIX]

	for _, j := range indices {
		// TODO: original code did a check: Node* n = _get(j); if (n) {...}
		n := idx.getNode(j)
		children = append(children, n)
	}

	children_indices := [2][]TIX{}
	data := make([]byte, idx.nodeSize) // Need it since, gc won't remove it until scope end

	m := idx.distance.MapNodeToMemory(
		unsafe.Pointer(unsafe.SliceData(data)), 0,
	)

	for attempt := 0; attempt < 3; attempt++ {
		children_indices[0] = nil
		children_indices[1] = nil

		idx.distance.CreateSplit(children, idx.nodeSize, idx.random, m)

		for _, j := range indices {
			// TODO: original code did a check: Node* n = _get(j); if (n) {...}
			n := idx.getNode(j)

			side := idx.distance.Side(
				m,
				n.GetVector(idx.vectorLength),
				idx.random,
			)

			children_indices[side] = append(children_indices[side], j)
		}

		if idx.splitImbalance(
			children_indices[0],
			children_indices[1]) < 0.95 {
			break
		}
	}

	threadedBuildPolicy.UnlockSharedNodes()

	// If we didn't find a hyperplane, just randomize sides as a last option
	for {
		if idx.splitImbalance(
			children_indices[interfaces.SideLeft],
			children_indices[interfaces.SideRight]) <= 0.99 {
			break
		}

		children_indices[0] = nil
		children_indices[1] = nil

		// Set the vector to 0.0
		m.SetVector(make([]TV, idx.vectorLength))

		for _, j := range indices {
			// Just randomize...
			side := idx.random.NextSide()
			children_indices[side] = append(children_indices[side], j)
		}
	}

	if isRoot {
		m.SetNumberOfDescendants(idx._n_items)
	} else {
		m.SetNumberOfDescendants(TIX(len(indices)))
	}

	var flip int
	if len(children_indices[interfaces.SideLeft]) > len(children_indices[interfaces.SideRight]) {
		flip = 1
	}

	child_first := make([]TIX, 2)

	for side := 0; side < 2; side++ {
		// run makeTree for the smallest child first (for cache locality)
		flip_side := side ^ flip

		child_first[flip_side] = idx.makeTree(
			children_indices[flip_side],
			false,
			rnd,
			threadedBuildPolicy,
		)
	}

	m.SetChildren(child_first)

	idx.buildPolicy.LockNNodes()
	idx.allocateSize(idx._n_nodes+1, threadedBuildPolicy)
	item := idx._n_nodes
	idx._n_nodes++
	idx.buildPolicy.UnlockNNodes()

	idx.buildPolicy.LockSharedNodes()
	dst := idx.getNode(item)

	utils.CopyNode(dst, m, idx.vectorLength)
	idx.buildPolicy.UnlockSharedNodes()

	if idx.logVerbose {
		fmt.Printf("added 2:node[item=%d] - %s\n", item, utils.DumpNode(idx.distance, dst))
	}

	return item
}

func (idx *AnnoyIndexImpl[TV, TIX]) splitImbalance(
	left_indices, right_indices []TIX) float64 {
	ls := float64(len(left_indices))
	rs := float64(len(right_indices))

	f := ls / (ls + rs + 1e-9) // Avoid 0/0
	return math.Max(f, 1-f)
}

func (idx *AnnoyIndexImpl[TV, TIX]) allocateSize(
	numNodes TIX,
	threadedBuildPolicy interfaces.AnnoyIndexBuildPolicy,
) {
	if numNodes > idx._nodes_size {

		if threadedBuildPolicy != nil {
			threadedBuildPolicy.LockNodes()
		}

		new_node_size := utils.Max(numNodes, TIX(float64(idx._nodes_size+1)*reallocation_factor))
		idx._nodes = idx.allocator.Reallocate(int(new_node_size * idx.nodeSize))
		idx._nodes_size = new_node_size

		if threadedBuildPolicy != nil {
			threadedBuildPolicy.UnlockNodes()
		}
	}
}
