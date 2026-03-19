package sort

import (
	"container/heap"

	"github.com/mariotoffia/goannoy/interfaces"
	"golang.org/x/exp/constraints"
)

type priorityQueueItems[T constraints.Ordered] []*interfaces.Pair[T]

func (pq priorityQueueItems[_]) Len() int {
	return len(pq)
}

func (pq priorityQueueItems[T]) Less(i, j int) bool {
	// Match std::priority_queue<pair<T, S>> from upstream Annoy: the largest
	// pair by (First, Second) priority is popped first.
	return pq[j].Less(pq[i])
}

func (pq priorityQueueItems[_]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueueItems[T]) Push(x interface{}) {
	*pq = append(*pq, x.(*interfaces.Pair[T]))
}

func (pq *priorityQueueItems[_]) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]
	return item
}

type PriorityQueue[T constraints.Ordered] struct {
	pq priorityQueueItems[T]
}

func NewPriorityQueue[T constraints.Ordered]() *PriorityQueue[T] {
	pq := make(priorityQueueItems[T], 0)
	heap.Init(&pq)

	return &PriorityQueue[T]{pq}
}

func (pq *PriorityQueue[_]) Len() int {
	return pq.pq.Len()
}

func (pq *PriorityQueue[_]) Empty() bool {
	return pq.Len() == 0
}

func (pq *PriorityQueue[T]) Push(first T, second interfaces.ItemID) {
	heap.Push(&pq.pq, &interfaces.Pair[T]{first, second})
}

func (pq *PriorityQueue[T]) Pop() *interfaces.Pair[T] {
	return heap.Pop(&pq.pq).(*interfaces.Pair[T])
}

func (pq *PriorityQueue[T]) Top() *interfaces.Pair[T] {
	if pq.Len() == 0 {
		return nil
	}
	return pq.pq[0]
}
