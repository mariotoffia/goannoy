package utils

import (
	"math"

	"github.com/mariotoffia/goannoy/interfaces"
)

func Max[T interfaces.IndexTypes](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Intersection[TIX interfaces.IndexTypes](a, b []TIX) []TIX {
	SortSlice(a)
	SortSlice(b)

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

func Round(f float64) float64 {
	return math.Floor(f + 0.5)
}

func RoundPlus(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return Round(f*shift) / shift
}
