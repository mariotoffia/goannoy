package index

import (
	"errors"
	"fmt"
	"math"
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/sort"
	"github.com/mariotoffia/goannoy/utils"
)

const reallocation_factor = float64(1.3)

// AnnoyIndexImpl is the actual index for all vectors.
//
// A Note from the authors https://github.com/spotify/annoy
//
// We use random projection to build a forest of binary trees of all items.
// Basically just split the hyperspace into two sides by a hyperplane,
// then recursively split each of those subtrees etc.
// We create a tree like this q times. The default q is determined automatically
// in such a way that we at most use 2x as much memory as the vectors take.
type AnnoyIndexImpl[TV interfaces.VectorType] struct {
	vectorLength int
	// nodeSize the the complete size of the node in bytes.
	nodeSize int
	// _n_items is how many nodes exists in the index.
	_n_items interfaces.ItemID
	_nodes   unsafe.Pointer
	_n_nodes interfaces.ItemID
	// _nodes_size is the number of nodes that has been allocated.
	// Total size is _node_size * nodeSize.
	_nodes_size int
	_roots      []interfaces.ItemID
	// batchMaxNNS is the maximum of indexes that a query can possibly create.
	// This is updated each time a index is loaded.
	batchMaxNNS          int
	logVerbose           bool
	maxDescendants       interfaces.ItemID
	random               interfaces.Random
	indexLoaded          bool
	indexBuilt           bool
	distance             interfaces.Distance[TV]
	buildPolicy          interfaces.AnnoyIndexBuildPolicy
	allocator            interfaces.BuildIndexAllocator
	indexMemoryAllocator interfaces.IndexAllocator
	indexMemory          interfaces.AllocatedIndex
	sorter               interfaces.Sorter[TV]
	onDiskFile           string
}

// New create a new index instance based on the _TV_ vector type. When done use
// the `io.Closer.Close()` to clean up any resources.
func New[TV interfaces.VectorType](
	random interfaces.Random,
	distance interfaces.Distance[TV],
	buildPolicy interfaces.AnnoyIndexBuildPolicy,
	allocator interfaces.BuildIndexAllocator,
	indexMemoryAllocator interfaces.IndexAllocator,
	sorter interfaces.Sorter[TV],
	logVerbose bool,
	hintNumIndexes int,
) interfaces.AnnoyIndex[TV] {
	if sorter == nil {
		sorter = &interfaces.SorterFunctions[TV]{
			SortSliceFunc:        sort.SortSlice,
			SortPairsFunc:        sort.SortPairs[TV],
			PartialSortSliceFunc: sort.PartialSortSlice[TV],
		}
	}

	index := &AnnoyIndexImpl[TV]{
		vectorLength:         distance.VectorLength(),
		random:               random,
		nodeSize:             distance.NodeSize(),
		maxDescendants:       distance.MaxNumChildren(),
		indexBuilt:           false,
		logVerbose:           logVerbose,
		distance:             distance,
		allocator:            allocator,
		buildPolicy:          buildPolicy,
		indexMemoryAllocator: indexMemoryAllocator,
		sorter:               sorter,
	}

	// Pre-allocate memory for the index if hintNumIndexes is set > 0.
	if hintNumIndexes > 0 {
		allocator.Reallocate(int(float64(distance.NodeSize()*hintNumIndexes) * reallocation_factor))
	}

	return index
}

// VectorLength returns the vector length of the index.
func (idx *AnnoyIndexImpl[TV]) VectorLength() int {
	return idx.vectorLength
}

func (idx *AnnoyIndexImpl[TV]) GetItem(itemIndex interfaces.ItemID) []TV {
	mustNonNegativeItemID("GetItem", itemIndex)

	v := idx.getNode(itemIndex).GetVector(idx.vectorLength)
	out := make([]TV, len(v))
	copy(out, v)
	return out
}

func (idx *AnnoyIndexImpl[TV]) GetNItems() int {
	return int(idx._n_items)
}

func (idx *AnnoyIndexImpl[TV]) GetNTrees() int {
	return len(idx._roots)
}

func (idx *AnnoyIndexImpl[TV]) AddItem(itemIndex interfaces.ItemID, v []TV) error {
	if idx.indexLoaded {
		return errors.New("can't add items to a loaded index")
	}

	if idx.indexBuilt {
		return errors.New("can't add items to a built index")
	}

	if itemIndex < 0 {
		return fmt.Errorf("negative item id: %d", itemIndex)
	}

	if idx.vectorLength != len(v) {
		return fmt.Errorf("vector length mismatch: %d != %d", idx.vectorLength, len(v))
	}

	// Ensure that we have enough memory for the new node.
	idx.allocateSize(int(itemIndex)+1, nil)

	// Map the node onto the memory.
	node := idx.getNode(itemIndex)

	// Initialize the node with the vector.
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

	return nil
}

func (idx *AnnoyIndexImpl[TV]) Build(numberOfTrees, numWorkers int) error {
	if idx.indexLoaded {
		return errors.New("can't build a loaded index")
	}

	if idx.indexBuilt {
		return errors.New("can't build a built index")
	}

	// Give the preprocessor a chance to process the nodes before building the index.
	idx.distance.PreProcess(idx._nodes, idx._n_items)

	idx._n_nodes = idx._n_items

	idx.buildPolicy.Build(idx, numberOfTrees, numWorkers)

	// Also, copy the roots into the last segment of the array.
	// This way we can load them faster without reading the whole file.
	idx.allocateSize(int(idx._n_nodes)+len(idx._roots), nil)

	for i := range idx._roots {
		dst := idx.getNode(idx._n_nodes + interfaces.ItemID(i))
		src := idx.getNode(idx._roots[i])

		utils.CopyNode(dst, src, idx.nodeSize)

		if idx.logVerbose {
			fmt.Printf(
				"added roots[i=%d]:%d node - %s\n", i, idx._roots[i], utils.DumpNode(idx.distance, dst),
			)
		}
	}

	idx._n_nodes += interfaces.ItemID(len(idx._roots))

	if pp, ok := idx.distance.(interface {
		PostProcess(nodes unsafe.Pointer, nodeCount interfaces.ItemID)
	}); ok {
		pp.PostProcess(idx._nodes, idx._n_items)
	}

	idx.indexBuilt = true

	// When building on disk, finalize the file and reload as read-only mmap.
	if idx.onDiskFile != "" {
		finalSize := int64(idx._n_nodes) * int64(idx.nodeSize)
		if trunc, ok := idx.allocator.(interface{ Truncate(int64) error }); ok {
			if err := trunc.Truncate(finalSize); err != nil {
				return fmt.Errorf("on-disk build: truncate: %w", err)
			}
		}
		idx.allocator.Free()
		return idx.Load(idx.onDiskFile)
	}

	idx.computeBatchMaxNNS()

	return nil
}

// ThreadBuild is called from the build policy to build the index.
func (idx *AnnoyIndexImpl[TV]) ThreadBuild(
	treesPerWorker, workerIdx int,
	threadedBuildPolicy interfaces.AnnoyIndexBuildPolicy,
) {
	rnd := idx.random.CloneAndReset()

	// Each worker needs its own seed, otherwise each worker would be building the same tree(s).
	rnd.SetSeed(rnd.GetSeed() + uint64(workerIdx))

	var threadRoots []interfaces.ItemID

	for {
		if treesPerWorker == -1 {
			threadedBuildPolicy.LockNNodes()
			if idx._n_nodes >= 2*idx._n_items {
				threadedBuildPolicy.UnlockNNodes()
				break
			}
			threadedBuildPolicy.UnlockNNodes()
		} else if len(threadRoots) >= treesPerWorker {
			break
		}

		var indices []interfaces.ItemID

		threadedBuildPolicy.LockSharedNodes()
		for i := 0; i < int(idx._n_items); i++ {
			itemID := interfaces.ItemID(i)
			node := idx.getNode(itemID)
			if node.GetNumberOfDescendants() >= 1 {
				indices = append(indices, itemID)
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

func (idx *AnnoyIndexImpl[TV]) getNode(index interfaces.ItemID) interfaces.Node[TV] {
	return idx.distance.MapNodeToMemory(idx._nodes, index)
}

func (idx *AnnoyIndexImpl[TV]) makeTree(
	indices []interfaces.ItemID, isRoot bool,
	rnd interfaces.Random,
	threadedBuildPolicy interfaces.AnnoyIndexBuildPolicy,
) interfaces.ItemID {
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

	lenIdx := interfaces.ItemID(len(indices))
	if lenIdx <= idx.maxDescendants &&
		(!isRoot || idx._n_items <= idx.maxDescendants || lenIdx == 1) {
		// Ensure we have memory for the new node.
		threadedBuildPolicy.LockNNodes()
		idx.allocateSize(int(idx._n_nodes)+1, threadedBuildPolicy)

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
			children := make([]interfaces.ItemID, len(indices))
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

	var children []interfaces.Node[TV]
	for _, j := range indices {
		// TODO: original code did a check: Node* n = _get(j); if (n) {...}
		n := idx.getNode(j)
		children = append(children, n)
	}

	childrenIndices := [2][]interfaces.ItemID{}
	data := make([]byte, idx.nodeSize) // Need it since gc won't remove it until scope end.
	m := idx.distance.MapNodeToMemory(
		unsafe.Pointer(unsafe.SliceData(data)),
		0,
	)

	for attempt := 0; attempt < 3; attempt++ {
		childrenIndices[0] = nil
		childrenIndices[1] = nil

		idx.distance.CreateSplit(children, idx.nodeSize, rnd, m)

		for _, j := range indices {
			// TODO: original code did a check: Node* n = _get(j); if (n) {...}
			n := idx.getNode(j)

			side := idx.distance.Side(
				m,
				n.GetVector(idx.vectorLength),
				rnd,
			)

			childrenIndices[side] = append(childrenIndices[side], j)
		}

		if idx.splitImbalance(childrenIndices[0], childrenIndices[1]) < 0.95 {
			break
		}
	}

	threadedBuildPolicy.UnlockSharedNodes()

	// If we didn't find a hyperplane, just randomize sides as a last option.
	for {
		if idx.splitImbalance(
			childrenIndices[interfaces.SideLeft],
			childrenIndices[interfaces.SideRight],
		) <= 0.99 {
			break
		}

		childrenIndices[0] = nil
		childrenIndices[1] = nil

		// Set the vector to 0.0.
		m.SetVector(make([]TV, idx.vectorLength))

		for _, j := range indices {
			// Just randomize...
			side := rnd.NextSide()
			childrenIndices[side] = append(childrenIndices[side], j)
		}
	}

	if isRoot {
		m.SetNumberOfDescendants(idx._n_items)
	} else {
		m.SetNumberOfDescendants(interfaces.ItemID(len(indices)))
	}

	flip := 0
	if len(childrenIndices[interfaces.SideLeft]) > len(childrenIndices[interfaces.SideRight]) {
		flip = 1
	}

	childFirst := make([]interfaces.ItemID, 2)
	for side := 0; side < 2; side++ {
		// run makeTree for the smallest child first (for cache locality)
		flipSide := side ^ flip
		childFirst[flipSide] = idx.makeTree(
			childrenIndices[flipSide],
			false,
			rnd,
			threadedBuildPolicy,
		)
	}

	m.SetChildren(childFirst)

	threadedBuildPolicy.LockNNodes()
	idx.allocateSize(int(idx._n_nodes)+1, threadedBuildPolicy)
	item := idx._n_nodes
	idx._n_nodes++
	threadedBuildPolicy.UnlockNNodes()

	threadedBuildPolicy.LockSharedNodes()
	dst := idx.getNode(item)
	utils.CopyNode(dst, m, idx.nodeSize)
	threadedBuildPolicy.UnlockSharedNodes()

	if idx.logVerbose {
		fmt.Printf("added 2:node[item=%d] - %s\n", item, utils.DumpNode(idx.distance, dst))
	}

	return item
}

func (idx *AnnoyIndexImpl[TV]) splitImbalance(
	leftIndices, rightIndices []interfaces.ItemID,
) float64 {
	ls := float64(len(leftIndices))
	rs := float64(len(rightIndices))

	f := ls / (ls + rs + 1e-9) // Avoid 0/0.
	return math.Max(f, 1-f)
}

// computeBatchMaxNNS calculates the maximum number of nearest neighbor
// candidates that can be collected during a single search traversal.
func (idx *AnnoyIndexImpl[TV]) computeBatchMaxNNS() {
	idx.batchMaxNNS = -1

	for i := 0; i < int(idx._n_nodes); i++ {
		nodeID := interfaces.ItemID(i)
		nd := idx.getNode(nodeID)

		nDescendants := nd.GetNumberOfDescendants()

		if nDescendants == 1 && nodeID < idx._n_items {
			idx.batchMaxNNS++
		} else if nDescendants <= idx.maxDescendants {
			idx.batchMaxNNS += len(nd.GetChildren())
		}
	}

	if idx.logVerbose {
		fmt.Println("Max NNS:", idx.batchMaxNNS)
	}
}

func (idx *AnnoyIndexImpl[TV]) allocateSize(
	numNodes int,
	threadedBuildPolicy interfaces.AnnoyIndexBuildPolicy,
) {
	if numNodes > idx._nodes_size {
		if threadedBuildPolicy != nil {
			threadedBuildPolicy.LockNodes()
		}

		newNodeSize := utils.Max(numNodes, int(float64(idx._nodes_size+1)*reallocation_factor))
		idx._nodes = idx.allocator.Reallocate(newNodeSize * idx.nodeSize)
		idx._nodes_size = newNodeSize

		if threadedBuildPolicy != nil {
			threadedBuildPolicy.UnlockNodes()
		}
	}
}

func mustNonNegativeItemID(method string, itemID interfaces.ItemID) {
	if itemID < 0 {
		panic(fmt.Sprintf("%s: negative item id %d", method, itemID))
	}
}
