package utils

import (
	"fmt"
	"testing"

	"github.com/mariotoffia/goannoy/random"
)

func BenchmarkSortPairsVsSortPairs2(t *testing.B) {
	i := 10
	for i <= 100000000 {
		testSet := createData(uint32(i))

		t.ResetTimer()

		t.Run(fmt.Sprintf("SortPairs/%d:", i), func(t *testing.B) {
			for i := 0; i < t.N; i++ {
				SortPairs(testSet)
			}
		})

		t.Run(fmt.Sprintf("SortPairs2/%d:", i), func(t *testing.B) {
			for i := 0; i < t.N; i++ {
				SortPairs2(testSet)
			}
		})

		if i < 1000000 {
			i *= 10
		} else if i >= 1000000 && i < 10000000 {
			i += 1000000
		} else {
			i *= 10
		}
	}
}

func BenchmarkSortVsSort2VsSort3(t *testing.B) {
	i := 10
	for i <= 100000000 {
		testSet := createIntData(uint32(i), 20)

		t.ResetTimer()

		t.Run(fmt.Sprintf("Sort/%d:", i), func(t *testing.B) {
			for i := 0; i < t.N; i++ {
				SortSlice(testSet)
			}
		})

		t.Run(fmt.Sprintf("Sort2/%d:", i), func(t *testing.B) {
			for i := 0; i < t.N; i++ {
				SortSlice2(testSet)
			}
		})

		t.Run(fmt.Sprintf("Sort3/%d:", i), func(t *testing.B) {
			for i := 0; i < t.N; i++ {
				SortSlice3(testSet)
			}
		})

		if i < 1000000 {
			i *= 10
		} else if i >= 1000000 && i < 10000000 {
			i += 1000000
		} else {
			i *= 10
		}
	}
}

func createIntData(N uint32, numSame int) []uint32 {
	rnd := random.NewGoRandom()
	s := make([]uint32, N)
	cntSame := 0
	last := rnd.NextIndex(N)

	for i := 0; i < len(s); i++ {
		if cntSame < numSame {
			s[i] = last
			cntSame++
		} else {
			s[i] = rnd.NextIndex(N)
			last = s[i]
			cntSame = 0
		}
	}
	return s
}
