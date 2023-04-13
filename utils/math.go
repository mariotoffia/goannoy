package utils

import (
	"github.com/jfcg/sorty/v2"
	"github.com/mariotoffia/goannoy/interfaces"
)

func Max[T interfaces.IndexTypes](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Intersection[TIX interfaces.IndexTypes](a, b []TIX) []TIX {
	sorty.SortSlice(a)
	sorty.SortSlice(b)

	maxSize := len(a)

	if len(b) > maxSize {
		maxSize = len(b)
	}

	// Pre-allocate the maximum possible size
	intersection := make([]TIX, maxSize)
	i, j, k := 0, 0, 0

	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			i++
		} else if a[i] > b[j] {
			j++
		} else {
			intersection[k] = a[i]
			i++
			j++
			k++
		}
	}

	return intersection[:k]
}
