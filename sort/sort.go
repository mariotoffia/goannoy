package sort

import (
	"sort"

	"github.com/jfcg/sorty/v2"
	"github.com/mariotoffia/goannoy/interfaces"
)

func SortSlice(slice []interfaces.ItemID) {
	sorty.SortSlice(slice)
}

func SortSlice2(slice []interfaces.ItemID) {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})
}

func SortSlice3(slice []interfaces.ItemID) {
	len := len(slice)

	if len > 500 {
		sort.Slice(slice, func(i, j int) bool {
			return slice[i] < slice[j]
		})
	} else {
		sorty.SortSlice(slice)
	}
}

func SortPairs[TV interfaces.VectorType](
	pairs []*interfaces.Pair[TV],
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

func SortPairs2[TV interfaces.VectorType](arr interfaces.Pairs[TV]) {
	n := arr.Len()

	for i := n/2 - 1; i >= 0; i-- {
		heapify(arr, n, i)
	}

	for i := n - 1; i >= 0; i-- {
		arr.Swap(0, i)
		heapify(arr, i, 0)
	}
}

func heapify[TV interfaces.VectorType](arr interfaces.Pairs[TV], n, i int) {
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
