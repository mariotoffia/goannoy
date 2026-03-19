package sort_test

import (
	"testing"

	"github.com/mariotoffia/goannoy/sort"
	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	pq := sort.NewPriorityQueue[float32]()

	pq.Push(1.1, 1)
	pq.Push(4.4, 4)
	pq.Push(3.3, 3)
	pq.Push(2.2, 2)

	expected := []struct {
		first  float32
		second int32
	}{
		{4.4, 4},
		{3.3, 3},
		{2.2, 2},
		{1.1, 1},
	}

	for i := 0; pq.Len() > 0; i++ {
		item := pq.Pop()

		assert.Equal(t, expected[i].first, item.First)
		assert.Equal(t, expected[i].second, item.Second)
	}
}
