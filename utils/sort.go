package utils

import (
	"sort"

	"github.com/mariotoffia/goannoy/interfaces"
)

func SortSlice[TIX interfaces.IndexTypes](slice []TIX) {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})
	//sorty.SortSlice(slice)
}

func SortPairs[TV interfaces.VectorType, TIX interfaces.IndexTypes](
	pairs []*Pair[TV, TIX],
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
