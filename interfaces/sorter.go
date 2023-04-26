package interfaces

import "golang.org/x/exp/constraints"

type Sorter[TV VectorType, TIX IndexTypes] interface {
	SortSlice(slice []TIX)
	SortPairs(pairs []*Pair[TV, TIX])
	PartialSortSlice(s Pairs[TV, TIX], begin, middle, end int)
}

type SorterFunctions[TV VectorType, TIX IndexTypes] struct {
	SortSliceFunc        func(slice []TIX)
	SortPairsFunc        func(pairs []*Pair[TV, TIX])
	PartialSortSliceFunc func(s Pairs[TV, TIX], begin, middle, end int)
}

func (sf *SorterFunctions[TV, TIX]) SortSlice(slice []TIX) {
	sf.SortSliceFunc(slice)
}

func (sf *SorterFunctions[TV, TIX]) SortPairs(pairs []*Pair[TV, TIX]) {
	sf.SortPairsFunc(pairs)
}

func (sf *SorterFunctions[TV, TIX]) PartialSortSlice(s Pairs[TV, TIX], begin, middle, end int) {
	sf.PartialSortSliceFunc(s, begin, middle, end)
}

type Pair[T constraints.Ordered, S constraints.Ordered] struct {
	First  T
	Second S
}

func (p *Pair[T, S]) Less(other *Pair[T, S]) bool {
	return p.First < other.First ||
		(p.First == other.First && p.Second < other.Second)
}

type Pairs[T constraints.Ordered, S constraints.Ordered] []*Pair[T, S]

func (pq Pairs[TV, TIX]) ContainsFirst(v TV) bool {
	for _, e := range pq {
		if e.First == v {
			return true
		}
	}
	return false
}

func (pq Pairs[TV, TIX]) ContainsSecond(v TIX) bool {
	for _, e := range pq {
		if e.Second == v {
			return true
		}
	}
	return false
}

func (pq Pairs[_, _]) Len() int {
	return len(pq)
}

func (pq Pairs[_, _]) Less(i, j int) bool {
	return pq[i].First < pq[j].First ||
		(pq[i].First == pq[j].First && pq[i].Second < pq[j].Second)
}

func (pq Pairs[_, _]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *Pairs[TP, TV]) Push(x interface{}) {
	*pq = append(*pq, x.(*Pair[TP, TV]))
}

func (pq *Pairs[_, _]) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]
	return item
}

func (pq *Pairs[TP, TV]) Top() *Pair[TP, TV] {
	if len(*pq) == 0 {
		return nil
	}

	return (*pq)[0]
}
