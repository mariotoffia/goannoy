package sort

import (
	"container/heap"

	"github.com/mariotoffia/goannoy/interfaces"
	"golang.org/x/exp/constraints"
)

type PriorityQueue[T constraints.Ordered, S constraints.Ordered] struct {
	pq interfaces.Pairs[T, S]
}

func NewPriorityQueue[T constraints.Ordered, S constraints.Ordered]() *PriorityQueue[T, S] {
	pq := make(interfaces.Pairs[T, S], 0)
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
	heap.Push(&pq.pq, &interfaces.Pair[T, S]{first, second})
}

func (pq *PriorityQueue[T, S]) Pop() *interfaces.Pair[T, S] {
	return heap.Pop(&pq.pq).(*interfaces.Pair[T, S])
}

func (pq *PriorityQueue[T, S]) Top() *interfaces.Pair[T, S] {
	if pq.Len() == 0 {
		return nil
	}
	return pq.pq[0]
}
