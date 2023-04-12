package utils

import (
	"container/heap"

	"golang.org/x/exp/constraints"
)

type PriorityQueue[T constraints.Ordered, S constraints.Ordered] struct {
	pq Pairs[T, S]
}

func NewPriorityQueue[T constraints.Ordered, S constraints.Ordered]() *PriorityQueue[T, S] {
	pq := make(Pairs[T, S], 0)
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
