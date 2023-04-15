package utils

import (
	"container/heap"

	"github.com/mariotoffia/goannoy/interfaces"
)

type pmnh []int

func (h pmnh) Len() int           { return len(h) }
func (h pmnh) Less(i, j int) bool { return h[i] < h[j] }
func (h pmnh) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *pmnh) Push(x interface{}) {
	*h = append(*h, x.(int))
}
func (h *pmnh) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func PartialSortSlice[TV interfaces.VectorType, TIX interfaces.IndexTypes](
	s []*Pair[TV, TIX],
	begin, middle, end int,
) {
	if begin >= end || middle <= begin || middle >= end {
		return
	}

	// Find the N smallest elements using a binary heap
	N := middle - begin
	h := make(pmnh, N)
	for i := 0; i < N; i++ {
		h[i] = i + begin
	}
	heap.Init(&h)
	for i := N; i < end-begin; i++ {
		if s[begin+i].Less(s[h[0]]) {
			h[0] = i + begin
			heap.Fix(&h, 0)
		}
	}

	// Swap elements
	for i := 0; i < N; i++ {
		s[begin+i], s[h[i]-begin] = s[h[i]-begin], s[begin+i]
	}

	// Sort sub-range [begin, middle) in place
	for i := begin + 1; i < middle; i++ {
		for j := i; j > begin && s[j].Less(s[j-1]); j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
