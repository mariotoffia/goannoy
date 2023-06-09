package angular_test

import (
	"testing"

	"github.com/mariotoffia/goannoy/builder"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/stretchr/testify/assert"
)

func createIndex(vectorLength int) interfaces.AnnoyIndex[float32, uint32] {
	return builder.Index[float32, uint32]().
		AngularDistance(vectorLength).
		SingleWorkerPolicy().
		Build()
}

func TestGetNnsByVectorReturnsCorrectIndexes(t *testing.T) {
	idx := createIndex(3)
	defer idx.Close()

	idx.AddItem(0, []float32{0, 0, 1})
	idx.AddItem(1, []float32{0, 1, 0})
	idx.AddItem(2, []float32{1, 0, 0})
	idx.Build(10, -1)

	ctx := idx.CreateContext()

	result, _ := idx.GetNnsByVector([]float32{3, 2, 1}, 3, -1, ctx)
	assert.Equal(t, []uint32{2, 1, 0}, result)

	result, _ = idx.GetNnsByVector([]float32{1, 2, 3}, 3, -1, ctx)
	assert.Equal(t, []uint32{0, 1, 2}, result)

	result, _ = idx.GetNnsByVector([]float32{2, 0, 1}, 3, -1, ctx)
	assert.Equal(t, []uint32{2, 0, 1}, result)
}

func TestGetNnsByItem(t *testing.T) {
	idx := createIndex(3)
	defer idx.Close()

	idx.AddItem(0, []float32{2, 1, 0})
	idx.AddItem(1, []float32{1, 2, 0})
	idx.AddItem(2, []float32{0, 0, 1})
	idx.Build(10, -1)

	ctx := idx.CreateContext()

	result, _ := idx.GetNnsByItem(0, 3, -1, ctx)
	assert.Equal(t, []uint32{0, 1, 2}, result)

	result, _ = idx.GetNnsByItem(1, 3, -1, ctx)
	assert.Equal(t, []uint32{1, 0, 2}, result)

}

func TestGetItem(t *testing.T) {
	idx := createIndex(3)
	defer idx.Close()

	idx.AddItem(0, []float32{2, 1, 0})
	idx.AddItem(1, []float32{1, 2, 0})
	idx.AddItem(2, []float32{0, 0, 1})
	idx.Build(10, -1)

	result := idx.GetItem(0)
	assert.Equal(t, []float32{2, 1, 0}, result)

	result = idx.GetItem(1)
	assert.Equal(t, []float32{1, 2, 0}, result)

	result = idx.GetItem(2)
	assert.Equal(t, []float32{0, 0, 1}, result)
}
