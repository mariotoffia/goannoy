package utils

import (
	"sort"

	"github.com/mariotoffia/goannoy/interfaces"
)

func SortSlice[TIX interfaces.IndexTypes](s []TIX) {
	// sorty.SortSlice(nns)
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}
