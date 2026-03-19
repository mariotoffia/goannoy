package index

import (
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/sort"
)

// BatchContext is a context that is used when calling `GetNnsByVector` and
// `GetNnsByItem`. Buffers grow dynamically to avoid silent candidate dropping.
type BatchContext[TV interfaces.VectorType] struct {
	nns      []interfaces.ItemID
	nns_dist []*interfaces.Pair[TV]
}

// CreateContext will create a batch context, that should be used in subsequent
// calls to `GetNnsByVector` and `GetNnsByItem`.
func (idx *AnnoyIndexImpl[TV]) CreateContext() interfaces.AnnoyIndexContext[TV] {
	initCap := idx.batchMaxNNS
	if initCap < 1 {
		initCap = int(idx._n_nodes) * 2
	}

	return &BatchContext[TV]{
		nns:      make([]interfaces.ItemID, 0, initCap),
		nns_dist: make([]*interfaces.Pair[TV], 0, initCap),
	}
}

// GetDistance returns the distance between the two indexes.
func (idx *AnnoyIndexImpl[TV]) GetDistance(i, j interfaces.ItemID) TV {
	mustNonNegativeItemID("GetDistance", i)
	mustNonNegativeItemID("GetDistance", j)

	ni := idx.distance.MapNodeToMemory(idx._nodes, i)
	nj := idx.distance.MapNodeToMemory(idx._nodes, j)

	return idx.distance.NormalizedDistance(
		idx.distance.Distance(ni, nj),
	)
}

// GetNnsByItem will search for the closest vectors to the given _item_ in the index.
// When _numNodesToInspect_ is -1, it will search number of trees in index * _numReturn_.
func (idx *AnnoyIndexImpl[TV]) GetNnsByItem(
	item interfaces.ItemID,
	numReturn, numNodesToInspect int,
	ctx interfaces.AnnoyIndexContext[TV],
) (result []interfaces.ItemID, distances []TV) {
	mustNonNegativeItemID("GetNnsByItem", item)

	node := idx.distance.MapNodeToMemory(idx._nodes, item)

	return idx.getNnsByVector(
		node.GetVector(idx.vectorLength),
		numReturn,
		numNodesToInspect,
		ctx,
		true,
	)
}

func (idx *AnnoyIndexImpl[TV]) GetNnsByItemIDs(
	item interfaces.ItemID,
	numReturn, numNodesToInspect int,
	ctx interfaces.AnnoyIndexContext[TV],
) []interfaces.ItemID {
	mustNonNegativeItemID("GetNnsByItem", item)

	node := idx.distance.MapNodeToMemory(idx._nodes, item)
	result, _ := idx.getNnsByVector(
		node.GetVector(idx.vectorLength),
		numReturn,
		numNodesToInspect,
		ctx,
		false,
	)

	return result
}

// GetNnsByVector will search for the closest vectors to the given _vector_.
// When _numNodesToInspect_ is -1, it will search number of trees in index * _numReturn_.
func (idx *AnnoyIndexImpl[TV]) GetNnsByVector(
	vector []TV,
	numReturn, numNodesToInspect int,
	ctx interfaces.AnnoyIndexContext[TV],
) (result []interfaces.ItemID, distances []TV) {
	return idx.getNnsByVector(vector, numReturn, numNodesToInspect, ctx, true)
}

func (idx *AnnoyIndexImpl[TV]) GetNnsByVectorIDs(
	vector []TV,
	numReturn, numNodesToInspect int,
	ctx interfaces.AnnoyIndexContext[TV],
) []interfaces.ItemID {
	result, _ := idx.getNnsByVector(vector, numReturn, numNodesToInspect, ctx, false)
	return result
}

func (idx *AnnoyIndexImpl[TV]) getNnsByVector(
	vector []TV,
	numReturn, numNodesToInspect int,
	ctx interfaces.AnnoyIndexContext[TV],
	includeDistances bool,
) (result []interfaces.ItemID, distances []TV) {
	bc := ctx.(*BatchContext[TV])
	q := sort.NewPriorityQueue[TV]()

	// Reset nns for reuse (keep underlying capacity).
	bc.nns = bc.nns[:0]

	if numNodesToInspect == -1 {
		numNodesToInspect = numReturn * len(idx._roots)
	}

	for i := range idx._roots {
		q.Push(idx.distance.PQInitialValue(), idx._roots[i])
	}

	cnt := 0

	for cnt < numNodesToInspect && !q.Empty() {
		top := q.Top()

		d := top.First
		i := top.Second
		nd := idx.distance.MapNodeToMemory(idx._nodes, i)

		q.Pop()

		nDescendants := nd.GetNumberOfDescendants()

		if nDescendants == 1 && i < idx._n_items {
			bc.nns = append(bc.nns, i)
			cnt++
		} else if nDescendants <= idx.maxDescendants {
			children := nd.GetChildren()
			n := int(nDescendants)
			bc.nns = append(bc.nns, children[:n]...)
			cnt += n
		} else {
			// Node is normal of the split plane.
			margin := idx.distance.Margin(nd, vector)
			children := nd.GetChildren()

			q.Push(
				idx.distance.PQDistance(d, margin, interfaces.SideRight),
				children[interfaces.SideRight],
			)

			q.Push(
				idx.distance.PQDistance(d, margin, interfaces.SideLeft),
				children[interfaces.SideLeft],
			)
		}
	}

	// Get distances for all items.
	// To avoid calculating distance multiple times for any items, sort by id.
	nns := bc.nns
	idx.sorter.SortSlice(nns)

	mem := make([]byte, idx.nodeSize) // Allocate mem on heap.

	// Prepare node to search for.
	vNode := idx.distance.MapNodeToMemory(
		unsafe.Pointer(unsafe.SliceData(mem)),
		0,
	)

	vNode.SetVector(vector)
	idx.distance.InitNode(vNode)

	var (
		lastset bool
		last    interfaces.ItemID
	)

	distCnt := 0

	for i := 0; i < len(nns); i++ {
		j := nns[i]
		if lastset && j == last {
			continue
		}

		last = j
		lastset = true
		n := idx.distance.MapNodeToMemory(idx._nodes, j)

		if n.GetNumberOfDescendants() == 1 { // This is only to guard a really obscure case, #284
			jn := idx.distance.MapNodeToMemory(idx._nodes, j)

			// Grow nns_dist if needed.
			if distCnt >= len(bc.nns_dist) {
				bc.nns_dist = append(bc.nns_dist, &interfaces.Pair[TV]{})
			}

			pair := bc.nns_dist[distCnt]
			pair.First = idx.distance.Distance(vNode, jn)
			pair.Second = j

			distCnt++
		}
	}

	nnsDist := bc.nns_dist[:distCnt]

	middle := distCnt
	if numReturn < distCnt {
		middle = numReturn
	}

	idx.sorter.PartialSortSlice(nnsDist, 0, middle, len(nnsDist))

	for i := 0; i < middle; i++ {
		result = append(result, nnsDist[i].Second)
		if includeDistances {
			distances = append(distances, idx.distance.NormalizedDistance(nnsDist[i].First))
		}
	}

	return
}
