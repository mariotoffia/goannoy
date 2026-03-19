package tests

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/mariotoffia/goannoy/builder"
	"github.com/mariotoffia/goannoy/interfaces"
	gorandom "github.com/mariotoffia/goannoy/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func angularIdx(dim int) interfaces.AnnoyIndex[float32] {
	return builder.Index().
		AngularDistance(dim).
		SingleWorkerPolicy().
		Build()
}

func angularIdxMmap(dim int) interfaces.AnnoyIndex[float32] {
	return builder.Index().
		AngularDistance(dim).
		SingleWorkerPolicy().
		MmapIndexAllocator().
		Build()
}

func angularIdxWithSeed(dim int, seed uint64) interfaces.AnnoyIndex[float32] {
	return builder.Index().
		Random(gorandom.NewKiss32Random(seed)).
		AngularDistance(dim).
		SingleWorkerPolicy().
		Build()
}

func angularIdxMmapWithSeed(dim int, seed uint64) interfaces.AnnoyIndex[float32] {
	return builder.Index().
		Random(gorandom.NewKiss32Random(seed)).
		AngularDistance(dim).
		SingleWorkerPolicy().
		MmapIndexAllocator().
		Build()
}

func angularIdxMultiWithSeed(dim int, seed uint64) interfaces.AnnoyIndex[float32] {
	return builder.Index().
		Random(gorandom.NewKiss32Random(seed)).
		AngularDistance(dim).
		UseMultiWorkerPolicy().
		MmapIndexAllocator().
		Build()
}

func upstreamBinaryFixturePath() string {
	return filepath.Join("testdata", "test.tree")
}

func requireFixture(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Skipf("fixture %s not available (run with .work/ checked out): %v", path, err)
	}
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

func assertContainsAll(t *testing.T, haystack []int32, needles []int32) {
	t.Helper()
	set := make(map[int32]bool)
	for _, v := range haystack {
		set[v] = true
	}
	for _, n := range needles {
		assert.True(t, set[n], "expected %d in result set", n)
	}
}
