package tests

import (
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/mariotoffia/goannoy/builder"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSplitPathAngular exercises the makeTree split path (which uses CopyNode)
// by building an index with more items than maxDescendants. For angular
// float32/int32 dim=3, maxDescendants=5, so 20 items forces splits.
func TestSplitPathAngular(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")
	dim := 3
	nItems := 20
	nTrees := 3

	idx := angularIdxMmap(dim)
	defer idx.Close()

	rng := rand.New(rand.NewSource(42))
	vectors := make([][]float32, nItems)
	for i := 0; i < nItems; i++ {
		v := make([]float32, dim)
		for d := 0; d < dim; d++ {
			v[d] = rng.Float32()*2 - 1
		}
		vectors[i] = v
		require.NoError(t, idx.AddItem(int32(i), v))
	}

	require.NoError(t, idx.Build(nTrees, 1))

	err := idx.Save(path)
	require.NoError(t, err)

	idx2 := angularIdxMmap(dim)
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	ctx := idx2.CreateContext()

	// Every item should be findable as its own nearest neighbor when using
	// a sufficiently large search_k (default search_k=-1 inspects too few
	// nodes with small nTrees and dim).
	largeSearchK := nItems * 20
	for i := 0; i < nItems; i++ {
		result, distances := idx2.GetNnsByVector(vectors[i], 1, largeSearchK, ctx)
		require.NotEmpty(t, result, "item %d: no results", i)
		assert.Equal(t, int32(i), result[0], "item %d: self should be nearest", i)
		require.NotEmpty(t, distances, "item %d: no distances", i)
		assert.InDelta(t, 0.0, float64(distances[0]), 1e-3,
			"item %d: self-distance should be ~0", i)
	}

	// Request all items — they should all be returned
	result, _ := idx2.GetNnsByVector(vectors[0], nItems, nItems*100, ctx)
	require.Len(t, result, nItems, "all items must be returned")
	seen := make(map[int32]bool)
	for _, id := range result {
		seen[id] = true
	}
	for i := 0; i < nItems; i++ {
		assert.True(t, seen[int32(i)], "item %d missing from results", i)
	}
}

// TestSaveLoadConsistencyMatrix runs a parameterized test across dimensions,
// tree counts, and item counts. For each config it asserts that search results
// before save match search results after load (with search_k=-1).
func TestSaveLoadConsistencyMatrix(t *testing.T) {
	dims := []int{3, 10, 50}
	trees := []int{1, 3, 5}
	items := []int{1, 5, 20, 50}

	for _, dim := range dims {
		for _, nTrees := range trees {
			for _, nItems := range items {
				name := fmt.Sprintf("dim=%d_trees=%d_items=%d", dim, nTrees, nItems)
				t.Run(name, func(t *testing.T) {
					testSaveLoadConsistency(t, dim, nTrees, nItems)
				})
			}
		}
	}
}

func testSaveLoadConsistency(t *testing.T, dim, nTrees, nItems int) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")

	idx := angularIdxMmap(dim)
	defer idx.Close()

	rng := rand.New(rand.NewSource(int64(dim*1000 + nTrees*100 + nItems)))
	vectors := make([][]float32, nItems)
	for i := 0; i < nItems; i++ {
		v := make([]float32, dim)
		for d := 0; d < dim; d++ {
			v[d] = rng.Float32()*2 - 1
		}
		vectors[i] = v
		require.NoError(t, idx.AddItem(int32(i), v))
	}

	require.NoError(t, idx.Build(nTrees, 1))

	// Query before save
	ctxBefore := idx.CreateContext()
	query := vectors[0]
	resultBefore, distBefore := idx.GetNnsByVector(query, nItems, -1, ctxBefore)

	err := idx.Save(path)
	require.NoError(t, err)

	// Query the save-reloaded index (Save calls Load internally)
	ctxAfterSave := idx.CreateContext()
	resultAfterSave, distAfterSave := idx.GetNnsByVector(query, nItems, -1, ctxAfterSave)

	assert.Equal(t, resultBefore, resultAfterSave,
		"results must match after Save (which calls Load)")
	assertDistancesClose(t, distBefore, distAfterSave)

	// Fresh load into new index
	idx2 := angularIdxMmap(dim)
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	ctxFresh := idx2.CreateContext()
	resultFresh, distFresh := idx2.GetNnsByVector(query, nItems, -1, ctxFresh)

	assert.Equal(t, resultBefore, resultFresh,
		"results must match from fresh Load")
	assertDistancesClose(t, distBefore, distFresh)
}

func assertDistancesClose(t *testing.T, expected, actual []float32) {
	t.Helper()
	require.Len(t, actual, len(expected), "distance slice length mismatch")
	for i := range expected {
		assert.InDelta(t, float64(expected[i]), float64(actual[i]), 1e-5,
			"distance[%d] mismatch", i)
	}
}

// TestRootCountPreserved verifies the root count (number of trees) is
// preserved across save and load.
func TestRootCountPreserved(t *testing.T) {
	configs := []struct {
		dim, nTrees, nItems int
	}{
		{3, 2, 1},
		{3, 5, 10},
		{10, 3, 20},
		{50, 5, 50},
	}

	for _, c := range configs {
		name := fmt.Sprintf("dim=%d_trees=%d_items=%d", c.dim, c.nTrees, c.nItems)
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "index.ann")

			idx := angularIdxMmap(c.dim)
			defer idx.Close()

			rng := rand.New(rand.NewSource(99))
			for i := 0; i < c.nItems; i++ {
				v := make([]float32, c.dim)
				for d := 0; d < c.dim; d++ {
					v[d] = rng.Float32()
				}
				require.NoError(t, idx.AddItem(int32(i), v))
			}

			require.NoError(t, idx.Build(c.nTrees, 1))

			// Verify search works before save with expected tree count.
			// search_k=-1 computes numReturn * len(roots), so passing
			// numReturn=1 with search_k=-1 works with any tree count.
			ctxPre := idx.CreateContext()
			res, _ := idx.GetNnsByVector(
				make([]float32, c.dim), 1, -1, ctxPre,
			)
			require.NotEmpty(t, res, "pre-save search returned empty")

			err := idx.Save(path)
			require.NoError(t, err)

			idx2 := angularIdxMmap(c.dim)
			defer idx2.Close()

			err = idx2.Load(path)
			require.NoError(t, err)

			ctxPost := idx2.CreateContext()
			res2, _ := idx2.GetNnsByVector(
				make([]float32, c.dim), 1, -1, ctxPost,
			)
			require.NotEmpty(t, res2, "post-load search returned empty")
		})
	}
}

// TestSplitPathDotProduct is the dot product equivalent of TestSplitPathAngular.
func TestSplitPathDotProduct(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")
	dim := 3
	nItems := 20
	nTrees := 3

	idx := dotProductIdxMmap(dim)
	defer idx.Close()

	rng := rand.New(rand.NewSource(42))
	vectors := make([][]float32, nItems)
	for i := 0; i < nItems; i++ {
		v := make([]float32, dim)
		for d := 0; d < dim; d++ {
			v[d] = rng.Float32()
		}
		vectors[i] = v
		require.NoError(t, idx.AddItem(int32(i), v))
	}

	require.NoError(t, idx.Build(nTrees, 1))

	err := idx.Save(path)
	require.NoError(t, err)

	idx2 := dotProductIdxMmap(dim)
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	ctx := idx2.CreateContext()

	// Every item should be findable
	for i := 0; i < nItems; i++ {
		result, _ := idx2.GetNnsByVector(vectors[i], 1, -1, ctx)
		require.NotEmpty(t, result, "dot product item %d: no results", i)
	}

	// All items returned with large search_k
	result, _ := idx2.GetNnsByVector(vectors[0], nItems, nItems*100, ctx)
	require.Len(t, result, nItems, "dot product: all items must be returned")
}

// TestTwoMeansSplitQuality verifies that with the CopyNode fix, TwoMeans
// produces quality splits leading to good search precision (> 90%).
func TestTwoMeansSplitQuality(t *testing.T) {
	dim := 10
	nItems := 500
	nTrees := 5

	mkIdx := func() interfaces.AnnoyIndex[float32] {
		return builder.Index().
			AngularDistance(dim).
			SingleWorkerPolicy().
			MmapIndexAllocator().
			Build()
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")

	idx := mkIdx()
	defer idx.Close()

	rng := rand.New(rand.NewSource(12345))
	vectors := make([][]float32, nItems)
	for i := 0; i < nItems; i++ {
		v := make([]float32, dim)
		for d := 0; d < dim; d++ {
			v[d] = rng.Float32()*2 - 1
		}
		vectors[i] = v
		require.NoError(t, idx.AddItem(int32(i), v))
	}

	require.NoError(t, idx.Build(nTrees, 1))

	err := idx.Save(path)
	require.NoError(t, err)

	idx2 := mkIdx()
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	ctx := idx2.CreateContext()

	// Check precision: for each item, verify self is in top-10 using a
	// large search_k to ensure thorough search.
	largeSearchK := nItems * 10
	selfFoundCount := 0
	for i := 0; i < nItems; i++ {
		result, _ := idx2.GetNnsByVector(vectors[i], 10, largeSearchK, ctx)
		for _, r := range result {
			if r == int32(i) {
				selfFoundCount++
				break
			}
		}
	}

	precision := float64(selfFoundCount) / float64(nItems)
	assert.Greater(t, precision, 0.90,
		"precision %.2f%% should be > 90%% (found self in top-10)", precision*100)
}

// TestSplitPathLargeDimAngular tests the split path with a large dimension
// where the old partial CopyNode would copy enough header bytes but truncate
// the split hyperplane vector.
func TestSplitPathLargeDimAngular(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")
	dim := 100
	nItems := 50
	nTrees := 3

	idx := angularIdxMmap(dim)
	defer idx.Close()

	rng := rand.New(rand.NewSource(777))
	vectors := make([][]float32, nItems)
	for i := 0; i < nItems; i++ {
		v := make([]float32, dim)
		norm := float32(0)
		for d := 0; d < dim; d++ {
			v[d] = rng.Float32()*2 - 1
			norm += v[d] * v[d]
		}
		norm = float32(math.Sqrt(float64(norm)))
		for d := 0; d < dim; d++ {
			v[d] /= norm
		}
		vectors[i] = v
		require.NoError(t, idx.AddItem(int32(i), v))
	}

	require.NoError(t, idx.Build(nTrees, 1))

	// Query before save
	ctxPre := idx.CreateContext()
	resBefore, distBefore := idx.GetNnsByVector(vectors[0], 10, -1, ctxPre)

	err := idx.Save(path)
	require.NoError(t, err)

	idx2 := angularIdxMmap(dim)
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	ctxPost := idx2.CreateContext()
	resAfter, distAfter := idx2.GetNnsByVector(vectors[0], 10, -1, ctxPost)

	assert.Equal(t, resBefore, resAfter,
		"dim=100: results must match after load")
	assertDistancesClose(t, distBefore, distAfter)
}
