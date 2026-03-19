package tests

import (
	"math/rand"
	"path/filepath"
	"sort"
	"testing"

	"github.com/mariotoffia/goannoy/builder"
	"github.com/mariotoffia/goannoy/interfaces"
	gorandom "github.com/mariotoffia/goannoy/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dotProductIdx(dim int) interfaces.AnnoyIndex[float32] {
	return builder.Index().
		DotProductDistance(dim).
		SingleWorkerPolicy().
		Build()
}

func dotProductIdxMmap(dim int) interfaces.AnnoyIndex[float32] {
	return builder.Index().
		DotProductDistance(dim).
		SingleWorkerPolicy().
		MmapIndexAllocator().
		Build()
}

var upstreamDotFixtureDataset = [][]float32{
	{-5, -6, -8},
	{2, -3, -3},
	{-2, 0, 2},
	{5, 3, 7},
	{1, 6, -5},
	{-3, -4, 0},
	{4, -1, 5},
	{0, 2, -7},
	{-4, 5, -2},
	{3, -5, 3},
	{-1, -2, 8},
	{-5, 1, -4},
	{2, 4, 1},
	{-2, -6, 6},
	{5, -3, -6},
	{1, 0, -1},
	{-3, 3, 4},
	{4, 6, -8},
	{0, -4, -3},
	{-4, -1, 2},
}

func upstreamDotFixturePath() string {
	return filepath.Join("testdata", "upstream_dot.tree")
}

func dotMetric(a, b []float32) float32 {
	var sum float32
	for i := range a {
		sum += a[i] * b[i]
	}
	return -sum
}

func recall(retrieved, relevant []int32) float64 {
	if len(relevant) == 0 {
		return 1
	}

	relevantSet := make(map[int32]struct{}, len(relevant))
	for _, id := range relevant {
		relevantSet[id] = struct{}{}
	}

	found := 0
	for _, id := range retrieved {
		if _, ok := relevantSet[id]; ok {
			found++
		}
	}

	return float64(found) / float64(len(relevantSet))
}

// Ported from Spotify Annoy test/dot_index_test.py.

func TestUpstreamDotGetNnsByVector(t *testing.T) {
	idx := dotProductIdx(2)
	defer idx.Close()

	idx.AddItem(0, []float32{2, 2})
	idx.AddItem(1, []float32{3, 2})
	idx.AddItem(2, []float32{3, 3})
	idx.Build(10, -1)

	ctx := idx.CreateContext()
	r, _ := idx.GetNnsByVector([]float32{4, 4}, 3, -1, ctx)
	assert.Equal(t, []int32{2, 1, 0}, r)

	r, _ = idx.GetNnsByVector([]float32{1, 1}, 3, -1, ctx)
	assert.Equal(t, []int32{2, 1, 0}, r)

	r, _ = idx.GetNnsByVector([]float32{4, 2}, 3, -1, ctx)
	assert.Equal(t, []int32{2, 1, 0}, r)
}

func TestUpstreamDotGetNnsByItem(t *testing.T) {
	idx := dotProductIdx(2)
	defer idx.Close()

	idx.AddItem(0, []float32{2, 2})
	idx.AddItem(1, []float32{3, 2})
	idx.AddItem(2, []float32{3, 3})
	idx.Build(10, -1)

	ctx := idx.CreateContext()
	r, _ := idx.GetNnsByItem(0, 3, -1, ctx)
	assert.Equal(t, []int32{2, 1, 0}, r)

	r, _ = idx.GetNnsByItem(2, 3, -1, ctx)
	assert.Equal(t, []int32{2, 1, 0}, r)
}

func TestUpstreamDotBinaryCompatibility(t *testing.T) {
	requireFixture(t, upstreamDotFixturePath())

	idx := dotProductIdxMmap(3)
	defer idx.Close()

	require.NoError(t, idx.Load(upstreamDotFixturePath()))

	assertVecAlmostEqual(t, upstreamDotFixtureDataset[0], idx.GetItem(0))
	assertVecAlmostEqual(t, upstreamDotFixtureDataset[19], idx.GetItem(19))
	assert.InDelta(t, 32.0, idx.GetDistance(0, 1), 1e-4)
	assert.InDelta(t, -6.0, idx.GetDistance(0, 2), 1e-4)
	assert.InDelta(t, -18.0, idx.GetDistance(3, 17), 1e-4)

	ctx := idx.CreateContext()
	result, distances := idx.GetNnsByItem(0, 5, 1000, ctx)
	assert.Equal(t, []int32{0, 11, 18, 7, 14}, result)
	require.Len(t, distances, 5)
	assert.InDeltaSlice(t, []float64{125, 51, 48, 44, 41}, []float64{
		float64(distances[0]),
		float64(distances[1]),
		float64(distances[2]),
		float64(distances[3]),
		float64(distances[4]),
	}, 1e-4)
}

func TestUpstreamDotBuildMatchesFixtureQueries(t *testing.T) {
	idx := builder.Index().
		Random(gorandom.NewKiss64Random(42)).
		DotProductDistance(3).
		SingleWorkerPolicy().
		Build()
	defer idx.Close()

	for i, v := range upstreamDotFixtureDataset {
		require.NoError(t, idx.AddItem(int32(i), v))
	}
	require.NoError(t, idx.Build(3, 1))

	ctx := idx.CreateContext()
	result0, _ := idx.GetNnsByItem(0, 5, 1000, ctx)
	assert.Equal(t, []int32{0, 11, 18, 7, 14}, result0)

	result3, _ := idx.GetNnsByItem(3, 5, 1000, ctx)
	assert.Equal(t, []int32{3, 6, 10, 12, 16}, result3)

	result7, _ := idx.GetNnsByItem(7, 5, 1000, ctx)
	assert.Equal(t, []int32{17, 7, 4, 0, 14}, result7)

	result19, _ := idx.GetNnsByItem(19, 5, 1000, ctx)
	assert.Equal(t, []int32{13, 10, 19, 16, 5}, result19)
}

func TestUpstreamDotDist(t *testing.T) {
	idx := dotProductIdx(2)
	defer idx.Close()

	idx.AddItem(0, []float32{0, 1})
	idx.AddItem(1, []float32{1, 1})
	idx.AddItem(2, []float32{0, 0})
	idx.Build(10, -1)

	assert.InDelta(t, 1.0, idx.GetDistance(0, 1), 1e-5)
	assert.InDelta(t, 0.0, idx.GetDistance(1, 2), 1e-5)
}

func dotRecallAt(t *testing.T, n, nTrees, nPoints, nRounds int) float64 {
	t.Helper()

	totalRecall := 0.0
	rng := rand.New(rand.NewSource(42))

	for r := 0; r < nRounds; r++ {
		const f = 10
		idx := dotProductIdx(f)

		data := make([][]float32, nPoints)
		for j := range data {
			v := make([]float32, f)
			for z := range v {
				v[z] = float32(rng.NormFloat64())
			}
			data[j] = v
			require.NoError(t, idx.AddItem(int32(j), v))
		}

		expectedResults := make([][]int32, nPoints)
		for i := range data {
			order := make([]int32, nPoints)
			for j := range order {
				order[j] = int32(j)
			}
			sort.Slice(order, func(a, b int) bool {
				left := dotMetric(data[i], data[int(order[a])])
				right := dotMetric(data[i], data[int(order[b])])
				if left == right {
					return order[a] < order[b]
				}
				return left < right
			})
			expectedResults[i] = order[:n]
		}

		idx.Build(nTrees, -1)
		ctx := idx.CreateContext()

		for i := range data {
			nns, _ := idx.GetNnsByVector(data[i], n, -1, ctx)
			totalRecall += recall(nns, expectedResults[i])
		}

		_ = idx.Close()
	}

	return totalRecall / float64(nRounds*nPoints)
}

func TestUpstreamDotRecallAt10(t *testing.T) {
	value := dotRecallAt(t, 10, 10, 1000, 5)
	assert.GreaterOrEqual(t, value, 0.65)
}

func TestUpstreamDotRecallAt100(t *testing.T) {
	value := dotRecallAt(t, 100, 10, 1000, 5)
	assert.GreaterOrEqual(t, value, 0.95)
}

func TestUpstreamDotRecallAt1000(t *testing.T) {
	value := dotRecallAt(t, 1000, 10, 1000, 5)
	assert.GreaterOrEqual(t, value, 0.99)
}

func TestUpstreamDotRecallAt1000FewerTrees(t *testing.T) {
	value := dotRecallAt(t, 1000, 4, 1000, 5)
	assert.GreaterOrEqual(t, value, 0.99)
}

func TestUpstreamDotGetNnsWithDistances(t *testing.T) {
	idx := dotProductIdx(3)
	defer idx.Close()

	idx.AddItem(0, []float32{0, 0, 2})
	idx.AddItem(1, []float32{0, 1, 1})
	idx.AddItem(2, []float32{1, 0, 0})
	idx.Build(10, -1)

	ctx := idx.CreateContext()
	l, d := idx.GetNnsByItem(0, 3, -1, ctx)
	assert.Equal(t, []int32{0, 1, 2}, l)
	require.Len(t, d, 3)
	assert.InDelta(t, 4.0, d[0], 1e-5)
	assert.InDelta(t, 2.0, d[1], 1e-5)
	assert.InDelta(t, 0.0, d[2], 1e-5)

	l, d = idx.GetNnsByVector([]float32{2, 2, 2}, 3, -1, ctx)
	assert.Equal(t, []int32{0, 1, 2}, l)
	require.Len(t, d, 3)
	assert.InDelta(t, 4.0, d[0], 1e-5)
	assert.InDelta(t, 4.0, d[1], 1e-5)
	assert.InDelta(t, 2.0, d[2], 1e-5)
}

func TestUpstreamDotIncludeDists(t *testing.T) {
	f := 40
	idx := dotProductIdx(f)
	defer idx.Close()

	v := make([]float32, f)
	rng := rand.New(rand.NewSource(123))
	var expectedDot float32
	for i := range v {
		v[i] = float32(rng.NormFloat64())
		expectedDot += v[i] * v[i]
	}

	negV := make([]float32, f)
	for i := range v {
		negV[i] = -v[i]
	}

	idx.AddItem(0, v)
	idx.AddItem(1, negV)
	idx.Build(10, -1)

	ctx := idx.CreateContext()
	indices, dists := idx.GetNnsByItem(0, 2, 10, ctx)
	assert.Equal(t, []int32{0, 1}, indices)
	require.Len(t, dists, 2)
	assert.InDelta(t, expectedDot, dists[0], 1e-3)
}

func TestUpstreamDotDistanceConsistency(t *testing.T) {
	const (
		n = 1000
		f = 3
	)

	idx := dotProductIdx(f)
	defer idx.Close()

	rng := rand.New(rand.NewSource(456))
	for j := 0; j < n; j++ {
		v := make([]float32, f)
		for z := range v {
			v[z] = float32(rng.NormFloat64())
		}
		require.NoError(t, idx.AddItem(int32(j), v))
	}
	require.NoError(t, idx.Build(10, -1))

	ctx := idx.CreateContext()
	sample := rand.New(rand.NewSource(789))
	for _, a := range sample.Perm(n)[:100] {
		indices, dists := idx.GetNnsByItem(int32(a), 100, -1, ctx)
		for i, b := range indices {
			u := idx.GetItem(int32(a))
			v := idx.GetItem(b)

			var dot float32
			for j := range u {
				dot += u[j] * v[j]
			}

			assert.InDelta(t, dot, dists[i], 1e-3)
			assert.InDelta(t, dot, idx.GetDistance(int32(a), b), 1e-3)
		}
	}
}

// Additional local regression coverage retained for save/load behavior.
func TestDotProductSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dp.ann")

	idx := dotProductIdxMmap(3)
	defer idx.Close()

	idx.AddItem(0, []float32{1.0, 0.0, 0.0})
	idx.AddItem(1, []float32{0.0, 1.0, 0.0})
	idx.AddItem(2, []float32{0.0, 0.0, 1.0})
	idx.Build(5, 1)

	err := idx.Save(path)
	require.NoError(t, err)

	idx2 := dotProductIdxMmap(3)
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	ctx := idx2.CreateContext()
	result, distances := idx2.GetNnsByVector([]float32{1.0, 0.0, 0.0}, 3, -1, ctx)

	require.Len(t, result, 3)
	assert.Equal(t, int32(0), result[0])
	require.Len(t, distances, 3)
}

func BenchmarkDotProductSearch(b *testing.B) {
	idx := dotProductIdx(3)
	defer idx.Close()

	for i := 0; i < 10; i++ {
		v := make([]float32, 3)
		v[i%3] = 1.0
		require.NoError(b, idx.AddItem(int32(i), v))
	}
	require.NoError(b, idx.Build(5, -1))

	ctx := idx.CreateContext()
	query := []float32{1.0, 0.0, 0.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.GetNnsByVector(query, 3, -1, ctx)
	}
}

func BenchmarkDotProductDistance(b *testing.B) {
	idx := dotProductIdx(3)
	defer idx.Close()

	idx.AddItem(0, []float32{1.0, 0.0, 0.0})
	idx.AddItem(1, []float32{0.0, 1.0, 0.0})
	idx.Build(5, -1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.GetDistance(0, 1)
	}
}
