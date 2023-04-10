package utils

import (
	"container/heap"

	"golang.org/x/exp/constraints"
)

type Pair[T constraints.Ordered, S constraints.Ordered] struct {
	First  T
	Second S
}

func (p *Pair[T, S]) Less(other *Pair[T, S]) bool {
	return p.First < other.First ||
		(p.First == other.First && p.Second < other.Second)
}

type PriorityQueue[T constraints.Ordered, S constraints.Ordered] struct {
	pq innerPQ[T, S]
}

func NewPriorityQueue[T constraints.Ordered, S constraints.Ordered]() *PriorityQueue[T, S] {
	pq := make(innerPQ[T, S], 0)
	heap.Init(&pq)

	return &PriorityQueue[T, S]{pq}
}

func (pq *PriorityQueue[_, _]) Len() int {
	return pq.pq.Len()
}

func (pq *PriorityQueue[_, _]) Empty() bool {
	return pq.Len() == 0
}

func (pq *PriorityQueue[T, S]) Push(first T, second S) {
	heap.Push(&pq.pq, &Pair[T, S]{first, second})
}

func (pq *PriorityQueue[T, S]) Pop() *Pair[T, S] {
	return heap.Pop(&pq.pq).(*Pair[T, S])
}

func (pq *PriorityQueue[T, S]) Top() *Pair[T, S] {
	if pq.Len() == 0 {
		return nil
	}
	return pq.pq[0]
}

type innerPQ[T constraints.Ordered, S constraints.Ordered] []*Pair[T, S]

func (pq innerPQ[_, _]) Len() int {
	return len(pq)
}

func (pq innerPQ[_, _]) Less(i, j int) bool {
	return pq[i].First < pq[j].First ||
		(pq[i].First == pq[j].First && pq[i].Second < pq[j].Second)
}

func (pq innerPQ[_, _]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *innerPQ[TP, TV]) Push(x interface{}) {
	*pq = append(*pq, x.(*Pair[TP, TV]))
}

func (pq *innerPQ[_, _]) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}
