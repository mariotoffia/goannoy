package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	pq := NewPriorityQueue[float32, int]()

	pq.Push(1.1, 1)
	pq.Push(4.4, 4)
	pq.Push(3.3, 3)
	pq.Push(2.2, 2)

	expectedSecond := int(1)
	expectFirst := float32(1.1)

	for pq.Len() > 0 {
		item := pq.Pop()

		assert.Equal(t, expectFirst, item.First)
		assert.Equal(t, expectedSecond, item.Second)

		expectedSecond++
		expectFirst += 1.1
		expectFirst = float32(int(expectFirst*10)) / 10
	}
}
