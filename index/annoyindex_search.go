package index

import (
	"sort"
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/utils"
)

// GetNnsByItem will search for the closest vectors to the given _item_ in the index. When
// _numReturn_ is -1, it will search number of trees in index * _numReturn_.
func (idx *AnnoyIndexImpl[TV, TIX]) GetNnsByItem(
	item TIX,
	numReturn, numNodesToInspect int,
) (result []TIX, distances []TV) {

	node := idx.distance.MapNodeToMemory(
		idx._nodes,
		item,
	)

	return idx.GetNnsByVector(
		node.GetVector(idx.vectorLength),
		numReturn,
		numNodesToInspect,
	)
}

// GetAllNns will search for the closest vectors to the given _vector_. When
// _numReturn_ is -1, it will search number of trees in index * _numReturn_.
func (idx *AnnoyIndexImpl[TV, TIX]) GetNnsByVector(
	vector []TV,
	numReturn, numNodesToInspect int,
) (result []TIX, distances []TV) {
	q := utils.NewPriorityQueue[TV, TIX]()

	if numNodesToInspect == -1 {
		numNodesToInspect = numReturn * len(idx._roots)
	}

	for i := range idx._roots {
		q.Push(idx.distance.PQInitialValue(), idx._roots[i])
	}

	nns := make([]TIX, 0, idx._n_nodes*2)

	for len(nns) < numNodesToInspect && !q.Empty() {
		top := q.Top()

		d := top.First
		i := top.Second
		nd := idx.distance.MapNodeToMemory(idx._nodes, i)

		q.Pop()

		nDescendants := nd.GetNumberOfDescendants()

		if nDescendants == 1 && i < idx._n_items {
			nns = append(nns, i)
		} else if nDescendants <= idx.maxDescendants {
			dst := nd.GetChildren()
			if len(dst) == int(nDescendants) {
				nns = append(nns, dst...)
			} else {
				nns = append(nns, dst[:nDescendants]...)
			}

		} else /*nDescendants > idx.maxDescendants*/ {
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
	utils.SortSlice(nns)

	mem := make([]byte, idx.nodeSize) // Allocate mem on gcheap

	v_node := idx.distance.MapNodeToMemory(
		unsafe.Pointer(unsafe.SliceData(mem)),
		0,
	)

	idx.distance.InitNode(v_node)

	nns_dist := make([]*utils.Pair[TV, TIX], len(nns))

	var (
		lastset bool
		last    TIX
		cnt     int
	)

	for i := 0; i < len(nns); i++ {
		j := nns[i]
		if lastset && j == last {
			continue
		}

		last = j
		n := idx.distance.MapNodeToMemory(idx._nodes, j)

		if n.GetNumberOfDescendants() == 1 { // This is only to guard a really obscure case, #284
			jn := idx.distance.MapNodeToMemory(idx._nodes, j)

			nns_dist[cnt] = &utils.Pair[TV, TIX]{
				First:  idx.distance.Distance(v_node, jn),
				Second: j,
			}

			cnt++
		}
	}

	if cnt < len(nns) {
		nns_dist = nns_dist[:cnt]
	}

	var p int
	if numReturn < cnt {
		p = numReturn
	} else {
		p = cnt
	}

	// Inefficient since it will sort the whole slice!
	sort.Slice(nns_dist, func(i, j int) bool {
		return nns_dist[i].Less(nns_dist[j])
	})

	nns_dist_partial := nns_dist[:p]

	for i := 0; i < len(nns_dist_partial); i++ {
		distances = append(distances, nns_dist_partial[i].First)
		result = append(result, nns_dist_partial[i].Second)
	}

	return
}
