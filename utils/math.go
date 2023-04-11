package utils

import "sort"

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Intersection(a, b []int) []int {
	sort.Ints(a)
	sort.Ints(b)

	maxSize := len(a)

	if len(b) > maxSize {
		maxSize = len(b)
	}

	// Pre-allocate the maximum possible size
	intersection := make([]int, maxSize)
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
