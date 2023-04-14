package utils

import (
	"sort"

	"github.com/jfcg/sorty/v2"
	"github.com/mariotoffia/goannoy/interfaces"
)

func SortSlice[TIX interfaces.IndexTypes](s []TIX) {
	sorty.SortSlice(s)
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}

func SortPairs[TV interfaces.VectorType, TIX interfaces.IndexTypes](
	nns_dist []*Pair[TV, TIX],
) {
	// Inefficient since it will sort the whole slice!
	/*sort.Slice(nns_dist, func(i, j int) bool {
		return nns_dist[i].Less(nns_dist[j])
	})*/

	lsw := func(i, k, r, s int) bool {
		if nns_dist[i].Less(nns_dist[k]) {
			if r != s {
				nns_dist[r], nns_dist[s] = nns_dist[s], nns_dist[r]
			}
			return true
		}
		return false
	}

	sorty.Sort(len(nns_dist), lsw)
}
