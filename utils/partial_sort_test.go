package utils

import (
	"fmt"
	"testing"

	"github.com/mariotoffia/goannoy/random"
	"github.com/stretchr/testify/assert"
)

func TestCorrectness(t *testing.T) {
	var s = []*Pair[float32, uint32]{
		{10, 10},
		{6, 6},
		{7, 7},
		{8, 8},
		{5, 5},
		{1, 1},
	}

	PartialSortSlice(s, 0, 3, len(s))

	assert.Equal(t, 6, len(s))
	assert.Equal(t, 1, int(s[0].First))
	assert.Equal(t, 5, int(s[1].First))
	assert.Equal(t, 6, int(s[2].First))

	contains := func(s []*Pair[float32, uint32], v float32) bool {
		for _, e := range s {
			if e.First == v {
				return true
			}
		}
		return false
	}

	assert.True(t, contains(s, 7))
	assert.True(t, contains(s, 8))
	assert.True(t, contains(s, 10))
}

func TestCorrectness2(t *testing.T) {
	var s = []*Pair[float32, uint32]{
		{10, 10},
		{6, 6},
		{7, 7},
		{8, 8},
		{5, 5},
		{1, 1},
	}

	PartialSortSlice2(s, 0, 3, len(s))

	assert.Equal(t, 6, len(s))
	assert.Equal(t, 1, int(s[0].First))
	assert.Equal(t, 5, int(s[1].First))
	assert.Equal(t, 6, int(s[2].First))

	contains := func(s []*Pair[float32, uint32], v float32) bool {
		for _, e := range s {
			if e.First == v {
				return true
			}
		}
		return false
	}

	assert.True(t, contains(s, 7))
	assert.True(t, contains(s, 8))
	assert.True(t, contains(s, 10))
}

func BenchmarkPartialSort(t *testing.B) {
	testSet := createData(uint32(1000000))
	N := len(testSet)

	t.Run("Partial", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			PartialSortSlice(testSet, 0, 10, N)
		}
	})

	t.Run("Partial2", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			PartialSortSlice2(testSet, 0, 10, N)
		}
	})

	t.Run("Sort", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			SortPairs(testSet)
		}
	})
}

func BenchmarkPartialSortVsSort(t *testing.B) {
	i := 10
	for i <= 100000000 {
		testSet := createData(uint32(i))
		N := len(testSet)

		t.ResetTimer()

		t.Run(fmt.Sprintf("Partial/%d", i), func(t *testing.B) {
			for i := 0; i < t.N; i++ {
				PartialSortSlice(testSet, 0, 10, N)
			}
		})

		t.Run(fmt.Sprintf("Sort/%d", i), func(t *testing.B) {
			for i := 0; i < t.N; i++ {
				SortPairs(testSet)
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

func createData(N uint32) []*Pair[float32, uint32] {
	rnd := random.NewGoRandom()
	s := make([]*Pair[float32, uint32], N)
	for i := 0; i < len(s); i++ {
		s[i] = &Pair[float32, uint32]{
			First:  float32(rnd.NextIndex(N)),
			Second: uint32(i),
		}
	}
	return s
}
