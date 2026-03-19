package interfaces

import "golang.org/x/exp/constraints"

type Sorter[TV VectorType] interface {
	SortSlice(slice []ItemID)
	SortPairs(pairs []*Pair[TV])
	PartialSortSlice(s Pairs[TV], begin, middle, end int)
}

type SorterFunctions[TV VectorType] struct {
	SortSliceFunc        func(slice []ItemID)
	SortPairsFunc        func(pairs []*Pair[TV])
	PartialSortSliceFunc func(s Pairs[TV], begin, middle, end int)
}

func (sf *SorterFunctions[TV]) SortSlice(slice []ItemID) {
	sf.SortSliceFunc(slice)
}

func (sf *SorterFunctions[TV]) SortPairs(pairs []*Pair[TV]) {
	sf.SortPairsFunc(pairs)
}

func (sf *SorterFunctions[TV]) PartialSortSlice(s Pairs[TV], begin, middle, end int) {
	sf.PartialSortSliceFunc(s, begin, middle, end)
}

type Pair[T constraints.Ordered] struct {
	First  T
	Second ItemID
}

func (p *Pair[T]) Less(other *Pair[T]) bool {
	return p.First < other.First ||
		(p.First == other.First && p.Second < other.Second)
}

type Pairs[T constraints.Ordered] []*Pair[T]

func (pq Pairs[TV]) ContainsFirst(v TV) bool {
	for _, e := range pq {
		if e.First == v {
			return true
		}
	}
	return false
}

func (pq Pairs[TV]) ContainsSecond(v ItemID) bool {
	for _, e := range pq {
		if e.Second == v {
			return true
		}
	}
	return false
}

func (pq Pairs[_]) Len() int {
	return len(pq)
}

func (pq Pairs[_]) Less(i, j int) bool {
	return pq[i].First < pq[j].First ||
		(pq[i].First == pq[j].First && pq[i].Second < pq[j].Second)
}

func (pq Pairs[_]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *Pairs[T]) Push(x interface{}) {
	*pq = append(*pq, x.(*Pair[T]))
}

func (pq *Pairs[_]) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]
	return item
}

func (pq *Pairs[T]) Top() *Pair[T] {
	if len(*pq) == 0 {
		return nil
	}

	return (*pq)[0]
}
