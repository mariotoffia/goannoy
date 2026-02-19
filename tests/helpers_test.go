package tests

import (
	"math"
	"testing"

	"github.com/mariotoffia/goannoy/builder"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func angularIdx(dim int) interfaces.AnnoyIndex[float32, uint32] {
	return builder.Index[float32, uint32]().
		AngularDistance(dim).
		SingleWorkerPolicy().
		Build()
}

func angularIdxMmap(dim int) interfaces.AnnoyIndex[float32, uint32] {
	return builder.Index[float32, uint32]().
		AngularDistance(dim).
		SingleWorkerPolicy().
		MmapIndexAllocator().
		Build()
}

func manualAngularDist(u, v []float32) float32 {
	var pp, qq, pq float64
	for i := range u {
		pp += float64(u[i]) * float64(u[i])
		qq += float64(v[i]) * float64(v[i])
		pq += float64(u[i]) * float64(v[i])
	}
	if pp == 0 || qq == 0 {
		return float32(math.Sqrt(2.0))
	}
	cosine := pq / math.Sqrt(pp*qq)
	if cosine > 1 {
		cosine = 1
	} else if cosine < -1 {
		cosine = -1
	}
	return float32(math.Sqrt(2.0 - 2.0*cosine))
}

func assertVecAlmostEqual(t *testing.T, expected, actual []float32) {
	t.Helper()
	require.Len(t, actual, len(expected))
	for i := range expected {
		assert.InDelta(t, expected[i], actual[i], 1e-3,
			"vector element %d mismatch", i)
	}
}

func assertContainsAll(t *testing.T, haystack []uint32, needles []uint32) {
	t.Helper()
	set := make(map[uint32]bool)
	for _, v := range haystack {
		set[v] = true
	}
	for _, n := range needles {
		assert.True(t, set[n], "expected %d in result set", n)
	}
}
