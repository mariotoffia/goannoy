package index

import (
	"sort"
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/utils"
)

func (idx *AnnoyIndexImpl[TV, TR]) Get_nns_by_item(
	item int,
	n, search_k int,
) (result []int, distances []TV) {

	node := idx.distance.MapNodeToMemory(
		idx._nodes,
		item,
		idx.vectorLength,
	)

	return idx.Get_all_nns(
		node.GetVector(idx.vectorLength),
		n,
		search_k,
	)
}

func (idx *AnnoyIndexImpl[TV, TR]) Get_nns_by_vector(
	vector []TV,
	n, search_k int,
) (result []int, distances []TV) {
	return idx.Get_all_nns(vector, n, search_k)
}

func (idx *AnnoyIndexImpl[TV, TR]) Get_all_nns(
	vector []TV,
	n, search_k int,
) (result []int, distances []TV) {
	q := utils.NewPriorityQueue[TV, int]()

	if search_k == -1 {
		search_k = len(idx._roots)
	}

	for i := range idx._roots[:search_k] {
		q.Push(idx.distance.PQInitialValue(), idx._roots[i])
	}

	nns := []int{}
	for len(nns) < search_k && !q.Empty() {
		top := q.Top()

		d := top.First
		i := top.Second
		nd := idx.distance.MapNodeToMemory(idx._nodes, i, idx.vectorLength)

		q.Pop()

		nDescendants := nd.GetNumberOfDescendants()

		if nDescendants == 1 && i < idx._n_items {
			nns = append(nns, i)
		} else if nDescendants < idx.maxDescendants {
			dst := nd.GetChildren()
			nns = append(nns, dst[:nDescendants]...)
		} else {
			margin := idx.distance.Margin(nd, vector, idx.vectorLength)
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
	sort.Ints(nns)

	mem := make([]byte, idx.nodeSize) // Allocate mem on gcheap

	v_node := idx.distance.MapNodeToMemory(
		unsafe.Pointer(unsafe.SliceData(mem)),
		0,
		idx.vectorLength,
	)

	v_node.InitNode(idx.vectorLength)

	nns_dist := []utils.Pair[TV, int]{}
	last := -1

	for i := 0; i < len(nns); i++ {
		j := nns[i]
		if j == last {
			continue
		}

		last = j
		n := idx.distance.MapNodeToMemory(idx._nodes, j, idx.vectorLength)

		if n.GetNumberOfDescendants() == 1 { // This is only to guard a really obscure case, #284
			jn := idx.distance.MapNodeToMemory(idx._nodes, j, idx.vectorLength)

			nns_dist = append(nns_dist, utils.Pair[TV, int]{
				First:  v_node.Distance(jn, idx.vectorLength),
				Second: j,
			})
		}
	}

	m := len(nns_dist)
	var p int
	if n < m {
		p = n
	} else {
		p = m
	}

	// Inefficient since it will sort the whole slice!
	sort.SliceStable(nns_dist, func(i, j int) bool {
		return nns_dist[i].Less(&nns_dist[j])
	})

	nns_dist_partial := nns_dist[:p]

	for i := 0; i < len(nns_dist_partial); i++ {
		distances = append(distances, nns_dist_partial[i].First)
		result = append(result, nns_dist_partial[i].Second)
	}

	return
}
