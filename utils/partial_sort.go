package utils

import (
	"container/heap"

	"github.com/mariotoffia/goannoy/interfaces"
)

func PartialSortSlice[TV interfaces.VectorType, TIX interfaces.IndexTypes](
	s []*Pair[TV, TIX],
	begin, middle, end int,
) {
	if begin >= end || middle <= begin || middle > end {
		return
	}

	// Find the N smallest elements
	N := middle - begin

	if end-begin > 20 && end-begin < 5000000 {
		// Use a heap
		SortPairs(s)
		return
	}

	for i := 0; i < N; i++ {
		minIndex := begin + i

		// Find the index of the smallest element in the unsorted part
		for j := begin + i + 1; j < end; j++ {
			if s[j].Less(s[minIndex]) {
				minIndex = j
			}
		}

		// Swap elements
		if minIndex != begin+i {
			s[begin+i], s[minIndex] = s[minIndex], s[begin+i]
		}
	}

	// Sort sub-range [begin, middle)
	if N > 15 {
		SortPairs(s[begin:middle])
	} else {
		for i := begin + 1; i < middle; i++ {
			for j := i; j > begin && s[j].Less(s[j-1]); j-- {
				s[j], s[j-1] = s[j-1], s[j]
			}
		}
	}
}

type pmnh[TV interfaces.VectorType, TIX interfaces.IndexTypes] struct {
	indices []int
	data    []*Pair[TV, TIX]
}

func (h pmnh[_, _]) Len() int           { return len(h.indices) }
func (h pmnh[_, _]) Less(i, j int) bool { return !h.data[h.indices[i]].Less(h.data[h.indices[j]]) }
func (h pmnh[_, _]) Swap(i, j int)      { h.indices[i], h.indices[j] = h.indices[j], h.indices[i] }
func (h *pmnh[TV, TIX]) Push(x interface{}) {
	*h = pmnh[TV, TIX]{
		indices: append(h.indices, x.(int)),
		data:    h.data,
	}
}
func (h *pmnh[_, _]) Pop() interface{} {
	old := h.indices
	n := len(old)
	x := old[n-1]
	h.indices = old[:n-1]
	return x
}

func PartialSortSlice2[TV interfaces.VectorType, TIX interfaces.IndexTypes](
	s []*Pair[TV, TIX],
	begin, middle, end int,
) {
	if begin >= end || middle <= begin || middle >= end {
		return
	}

	// Find the N smallest elements using a binary heap
	N := middle - begin
	h := pmnh[TV, TIX]{indices: make([]int, N), data: s}
	for i := 0; i < N; i++ {
		h.indices[i] = i + begin
	}
	heap.Init(&h)
	for i := N; i < end-begin; i++ {
		if s[begin+i].Less(s[h.indices[0]]) {
			h.indices[0] = i + begin
			heap.Fix(&h, 0)
		}
	}

	// Swap elements
	for i := 0; i < N; i++ {
		s[begin+i], s[h.indices[i]-begin] = s[h.indices[i]-begin], s[begin+i]
	}

	// Sort sub-range [begin, middle) in place
	for i := begin + 1; i < middle; i++ {
		for j := i; j > begin && s[j].Less(s[j-1]); j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
