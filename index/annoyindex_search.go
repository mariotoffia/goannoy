package index

import (
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/utils"
)

// BatchContext is a context that is used when calling `GetNnsByVector` and
// `GetNnsByItem`.
type BatchContext[TV interfaces.VectorType, TIX interfaces.IndexTypes] struct {
	nns      []TIX
	nns_dist []*utils.Pair[TV, TIX]
	length   int
}

// CreateContext will create a batch context, that should be used in subsequent
// calls to `GetNnsByVector` and `GetNnsByItem`.
func (idx *AnnoyIndexImpl[TV, TIX]) CreateContext() interfaces.AnnoyIndexContext[TV, TIX] {
	nnsLen := idx.batchMaxNNS
	if nnsLen < 1 {
		nnsLen = int(idx._n_nodes) * 2
	}

	bc := &BatchContext[TV, TIX]{
		length:   nnsLen,
		nns:      make([]TIX, nnsLen),
		nns_dist: make([]*utils.Pair[TV, TIX], nnsLen),
	}

	for i := 0; i < nnsLen; i++ {
		bc.nns_dist[i] = &utils.Pair[TV, TIX]{
			First:  0,
			Second: 0,
		}
	}

	return bc
}

// GetNnsByItem will search for the closest vectors to the given _item_ in the index. When
// _numReturn_ is -1, it will search number of trees in index * _numReturn_.
func (idx *AnnoyIndexImpl[TV, TIX]) GetNnsByItem(
	item TIX,
	numReturn, numNodesToInspect int,
	ctx interfaces.AnnoyIndexContext[TV, TIX],
) (result []TIX, distances []TV) {

	node := idx.distance.MapNodeToMemory(
		idx._nodes,
		item,
	)

	return idx.GetNnsByVector(
		node.GetVector(idx.vectorLength),
		numReturn,
		numNodesToInspect,
		ctx,
	)
}

// GetAllNns will search for the closest vectors to the given _vector_. When
// _numReturn_ is -1, it will search number of trees in index * _numReturn_.
func (idx *AnnoyIndexImpl[TV, TIX]) GetNnsByVector(
	vector []TV,
	numReturn, numNodesToInspect int,
	ctx interfaces.AnnoyIndexContext[TV, TIX],
) (result []TIX, distances []TV) {
	bc := ctx.(*BatchContext[TV, TIX])
	q := utils.NewPriorityQueue[TV, TIX]()

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
			bc.nns[cnt] = i
			cnt++
		} else if nDescendants <= idx.maxDescendants {
			dst := nd.GetChildren()
			if len(dst) == int(nDescendants) {
				copy(bc.nns[TIX(cnt):], dst)
			} else {
				copy(bc.nns[TIX(cnt):], dst[:nDescendants])
			}
			cnt += int(nDescendants)
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

	// Get distances for all items
	// To avoid calculating distance multiple times for any items, sort by id
	nns := bc.nns[:cnt]
	utils.SortSlice(nns)

	mem := make([]byte, idx.nodeSize) // Allocate mem on gcheap

	// Prepare node to search for
	v_node := idx.distance.MapNodeToMemory(
		unsafe.Pointer(unsafe.SliceData(mem)),
		0,
	)

	v_node.SetVector(vector)

	idx.distance.InitNode(v_node)

	var (
		lastset bool
		last    TIX
	)

	cnt = 0

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

			pair := bc.nns_dist[cnt]
			pair.First = idx.distance.Distance(v_node, jn)
			pair.Second = j

			cnt++
		}
	}

	var nns_dist []*utils.Pair[TV, TIX]
	if cnt < len(nns) {
		nns_dist = bc.nns_dist[:cnt]
	} else {
		nns_dist = bc.nns_dist[:len(nns)]
	}

	var middle int
	if numReturn < cnt {
		middle = numReturn
	} else {
		middle = cnt
	}

	// Inefficient since it will sort the whole slice!
	utils.PartialSortSlice(nns_dist, 0, middle, len(nns_dist))

	//nns_dist_partial := nns_dist[:middle]

	for i := 0; i < middle; i++ {
		distances = append(distances, nns_dist[i].First)
		result = append(result, nns_dist[i].Second)
	}

	return
}
