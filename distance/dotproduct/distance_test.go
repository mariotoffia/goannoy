package dotproduct

import (
	"math"
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

func TestGetDotProductDistance(t *testing.T) {
	idx := createIndex(2)
	defer idx.Close()

	idx.AddItem(0, []float32{0, 1})
	idx.AddItem(1, []float32{1, 1})
	idx.Build(10, -1)

	dst := idx.GetDistance(0, 1)
	zero := math.Abs(float64(1.0 - dst))

	assert.Equal(t, float32(1), dst)
	assert.Equal(t, 0.0, zero)
	assert.True(t, zero < 0.00001)

}
