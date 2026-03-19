package tests

import (
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNoCandidateDropping verifies that increasing search_k returns at least
// as many unique results. Under the old fixed-buffer scheme, candidates could
// be silently dropped when the buffer filled up.
func TestNoCandidateDropping(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")
	dim := 10
	nItems := 200
	nTrees := 5

	idx := angularIdxMmap(dim)
	defer idx.Close()

	rng := rand.New(rand.NewSource(42))
	query := make([]float32, dim)
	for i := 0; i < nItems; i++ {
		v := make([]float32, dim)
		for d := 0; d < dim; d++ {
			v[d] = rng.Float32()*2 - 1
		}
		if i == 0 {
			copy(query, v)
		}
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

	// Increasing search_k should yield >= as many unique results
	searchKs := []int{10, 50, 200, 1000}
	prevCount := 0
	for _, sk := range searchKs {
		result, _ := idx2.GetNnsByVector(query, nItems, sk, ctx)
		uniqueCount := countUnique(result)
		assert.GreaterOrEqual(t, uniqueCount, prevCount,
			"search_k=%d should return >= unique results than search_k=%d",
			sk, searchKs[0])
		prevCount = uniqueCount
	}
}

func countUnique(ids []int32) int {
	seen := make(map[int32]bool, len(ids))
	for _, id := range ids {
		seen[id] = true
	}
	return len(seen)
}

// TestSearchKDefaultVsExplicit verifies that search_k=-1 (default) produces
// results that are a subset of what a much larger explicit search_k returns.
func TestSearchKDefaultVsExplicit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")
	dim := 10
	nItems := 100
	nTrees := 3

	idx := angularIdxMmap(dim)
	defer idx.Close()

	rng := rand.New(rand.NewSource(123))
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

	// Default search_k (-1) results
	resultDefault, _ := idx2.GetNnsByVector(vectors[0], 10, -1, ctx)

	// Large explicit search_k results
	resultLarge, _ := idx2.GetNnsByVector(vectors[0], 10, nItems*10, ctx)

	// The default result's top-1 should also appear in the large result
	require.NotEmpty(t, resultDefault)
	require.NotEmpty(t, resultLarge)

	largeSet := make(map[int32]bool)
	for _, id := range resultLarge {
		largeSet[id] = true
	}
	assert.True(t, largeSet[resultDefault[0]],
		"default top-1 result %d should be in large search_k results", resultDefault[0])
}

// TestContextReuse verifies that a BatchContext can be safely reused across
// multiple queries without data corruption.
func TestContextReuse(t *testing.T) {
	dim := 5
	nItems := 30
	nTrees := 3

	idx := angularIdxMmap(dim)
	defer idx.Close()

	rng := rand.New(rand.NewSource(99))
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

	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")
	err := idx.Save(path)
	require.NoError(t, err)

	idx2 := angularIdxMmap(dim)
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	ctx := idx2.CreateContext()

	// Run the same query 10 times — results must be identical
	var firstResult []int32
	var firstDists []float32
	for iter := 0; iter < 10; iter++ {
		result, dists := idx2.GetNnsByVector(vectors[0], 5, -1, ctx)
		if iter == 0 {
			firstResult = result
			firstDists = dists
		} else {
			assert.Equal(t, firstResult, result,
				"iteration %d: results differ from first query", iter)
			assertDistancesClose(t, firstDists, dists)
		}
	}
}
