package utils

import (
	"github.com/jfcg/sorty/v2"
	"github.com/mariotoffia/goannoy/interfaces"
)

// SortSlice sorts the slice of TIX
func SortSlice[TIX interfaces.IndexTypes](a []TIX) {
	sorty.SortSlice(a)
	/*
		sort.Slice(a, func(i, j int) bool {
			return a[i] < a[j]
		})
	*/
}
