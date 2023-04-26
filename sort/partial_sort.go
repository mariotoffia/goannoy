package sort

import (
	"github.com/mariotoffia/goannoy/interfaces"
)

func PartialSortSlice[TV interfaces.VectorType, TIX interfaces.IndexTypes](
	s interfaces.Pairs[TV, TIX],
	begin, middle, end int,
) {
	if begin >= end || middle <= begin || middle > end {
		return
	}

	// Find the N smallest elements
	N := middle - begin

	if end-begin > 20 && end-begin < 5000000 {
		SortPairs(s)
		return
	}

	for i := 0; i < N; i++ {
		minIndex := begin + i

		// Find the index of the smallest element in the unsorted part
		for j := begin + i + 1; j < end; j++ {
			if s[j].Less(s[minIndex]) {
				minIndex = j
			}
		}

		// Swap elements
		if minIndex != begin+i {
			s[begin+i], s[minIndex] = s[minIndex], s[begin+i]
		}
	}

	// Sort sub-range [begin, middle)
	if N > 15 {
		SortPairs(s[begin:middle])
	} else {
		for i := begin + 1; i < middle; i++ {
			for j := i; j > begin && s[j].Less(s[j-1]); j-- {
				s[j], s[j-1] = s[j-1], s[j]
			}
		}
	}
}

func PartialSortSlice2[TV interfaces.VectorType, TIX interfaces.IndexTypes](
	s interfaces.Pairs[TV, TIX],
	begin, middle, end int,
) {
	beginMiddle := middle - begin

	if beginMiddle >= s.Len() {
		SortPairs(s)
		return
	}

	if begin >= end || middle <= begin || middle > end {
		return
	}

	buildMaxHeap(s[begin:middle], beginMiddle)

	for i := middle; i < end; i++ {
		if s.Less(i, begin) {
			s.Swap(begin, i)
			maxHeapify(s[begin:middle], 0, beginMiddle)
		}
	}

	SortPairs2(s[begin:middle])

}

func buildMaxHeap[TV interfaces.VectorType, TIX interfaces.IndexTypes](arr interfaces.Pairs[TV, TIX], n int) {
	for i := n/2 - 1; i >= 0; i-- {
		maxHeapify(arr, i, n)
	}
}

func maxHeapify[TV interfaces.VectorType, TIX interfaces.IndexTypes](arr interfaces.Pairs[TV, TIX], i, n int) {
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
		maxHeapify(arr, largest, n)
	}
}
