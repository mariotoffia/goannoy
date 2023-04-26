package sort

import (
	"sort"

	"github.com/jfcg/sorty/v2"
	"github.com/mariotoffia/goannoy/interfaces"
)

func SortSlice[TIX interfaces.IndexTypes](slice []TIX) {
	sorty.SortSlice(slice)
}

func SortSlice2[TIX interfaces.IndexTypes](slice []TIX) {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})
}

func SortSlice3[TIX interfaces.IndexTypes](slice []TIX) {
	len := len(slice)

	if len > 500 {
		sort.Slice(slice, func(i, j int) bool {
			return slice[i] < slice[j]
		})
	} else {
		sorty.SortSlice(slice)
	}
}

func SortPairs[TV interfaces.VectorType, TIX interfaces.IndexTypes](
	pairs []*interfaces.Pair[TV, TIX],
) {
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Less(pairs[j])
	})
	/*
	   	lsw := func(i, k, r, s int) bool {
	   		if pairs[i].Less(pairs[k]) {
	   			if r != s {
	   				pairs[r], pairs[s] = pairs[s], pairs[r]
	   			}
	   			return true
	   		}
	   		return false
	   	}

	   sorty.Sort(len(pairs), lsw)
	*/
}

func SortPairs2[TV interfaces.VectorType, TIX interfaces.IndexTypes](arr interfaces.Pairs[TV, TIX]) {
	n := arr.Len()

	for i := n/2 - 1; i >= 0; i-- {
		heapify(arr, n, i)
	}

	for i := n - 1; i >= 0; i-- {
		arr.Swap(0, i)
		heapify(arr, i, 0)
	}
}

func heapify[TV interfaces.VectorType, TIX interfaces.IndexTypes](arr interfaces.Pairs[TV, TIX], n, i int) {
	largest := i
	left := 2*i + 1
	right := 2*i + 2

	if left < n && arr.Less(largest, left) {
		largest = left
	}

	if right < n && arr.Less(largest, right) {
		largest = right
	}

	if largest != i {
		arr.Swap(i, largest)
		heapify(arr, n, largest)
	}
}
