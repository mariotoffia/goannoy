package tests

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/mariotoffia/goannoy/builder"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOneItemBuildSearchAngular verifies that searching a 1-item index
// works without Save/Load.
func TestOneItemBuildSearchAngular(t *testing.T) {
	idx := builder.Index().
		AngularDistance(3).
		SingleWorkerPolicy().
		Build()
	defer idx.Close()

	idx.AddItem(0, []float32{1.0, 0.0, 0.0})
	idx.Build(2, 1)

	ctx := idx.CreateContext()
	result, distances := idx.GetNnsByVector([]float32{1.0, 0.0, 0.0}, 1, -1, ctx)

	require.Len(t, result, 1)
	assert.Equal(t, int32(0), result[0])
	require.Len(t, distances, 1)
}

// TestOneItemSaveLoadSearchAngular is the exact reproducer for issue #5:
// GetNnsByVector panics with slice bounds out of range after Load on a
// small index.
func TestOneItemSaveLoadSearchAngular(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")

	idx := builder.Index().
		AngularDistance(3).
		SingleWorkerPolicy().
		MmapIndexAllocator().
		Build()
	defer idx.Close()

	idx.AddItem(0, []float32{0.5, 0.5, 0.0})
	idx.Build(2, 1)

	err := idx.Save(path)
	require.NoError(t, err)

	idx2 := builder.Index().
		AngularDistance(3).
		SingleWorkerPolicy().
		MmapIndexAllocator().
		Build()
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	ctx := idx2.CreateContext()
	result, distances := idx2.GetNnsByVector([]float32{0.5, 0.5, 0.0}, 1, -1, ctx)

	require.Len(t, result, 1)
	assert.Equal(t, int32(0), result[0])
	require.Len(t, distances, 1)
}

// TestTwoItemSaveLoadSearchAngular tests a 2-item index through Save/Load/Search.
func TestTwoItemSaveLoadSearchAngular(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")

	idx := builder.Index().
		AngularDistance(3).
		SingleWorkerPolicy().
		MmapIndexAllocator().
		Build()
	defer idx.Close()

	idx.AddItem(0, []float32{1.0, 0.0, 0.0})
	idx.AddItem(1, []float32{0.0, 1.0, 0.0})
	idx.Build(2, 1)

	err := idx.Save(path)
	require.NoError(t, err)

	idx2 := builder.Index().
		AngularDistance(3).
		SingleWorkerPolicy().
		MmapIndexAllocator().
		Build()
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	ctx := idx2.CreateContext()
	result, distances := idx2.GetNnsByVector([]float32{1.0, 0.0, 0.0}, 2, -1, ctx)

	require.Len(t, result, 2)
	assert.Equal(t, int32(0), result[0], "closest item should be 0")
	require.Len(t, distances, 2)
}

// TestSmallIndexVariousConfigs runs parameterized tests across different
// item counts and tree counts to ensure no panics or incorrect results.
func TestSmallIndexVariousConfigs(t *testing.T) {
	itemCounts := []int{1, 2, 3, 5}
	treeCounts := []int{1, 2, 3, 5}

	for _, nItems := range itemCounts {
		for _, nTrees := range treeCounts {
			name := fmt.Sprintf("items=%d_trees=%d", nItems, nTrees)
			t.Run(name, func(t *testing.T) {
				testSmallIndexSaveLoadSearch(t, nItems, nTrees)
			})
		}
	}
}

func testSmallIndexSaveLoadSearch(t *testing.T, nItems, nTrees int) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")
	dim := 3

	idx := builder.Index().
		AngularDistance(dim).
		SingleWorkerPolicy().
		MmapIndexAllocator().
		Build()
	defer idx.Close()

	vectors := make([][]float32, nItems)
	for i := 0; i < nItems; i++ {
		v := make([]float32, dim)
		v[i%dim] = 1.0
		vectors[i] = v
		require.NoError(t, idx.AddItem(int32(i), v))
	}

	require.NoError(t, idx.Build(nTrees, 1))

	err := idx.Save(path)
	require.NoError(t, err)

	idx2 := builder.Index().
		AngularDistance(dim).
		SingleWorkerPolicy().
		MmapIndexAllocator().
		Build()
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	ctx := idx2.CreateContext()

	for i := 0; i < nItems; i++ {
		result, distances := idx2.GetNnsByVector(vectors[i], 1, -1, ctx)

		require.NotEmpty(t, result,
			"expected at least one result for item %d", i)
		require.NotEmpty(t, distances,
			"expected at least one distance for item %d", i)
	}
}

// TestSearchResultsMatchAfterLoad verifies that search results from a
// loaded index match those from the originally built index.
func TestSearchResultsMatchAfterLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")
	dim := 3

	mkIndex := func() interfaces.AnnoyIndex[float32] {
		return builder.Index().
			AngularDistance(dim).
			SingleWorkerPolicy().
			MmapIndexAllocator().
			Build()
	}

	idx := mkIndex()
	defer idx.Close()

	idx.AddItem(0, []float32{1.0, 0.0, 0.0})
	idx.AddItem(1, []float32{0.0, 1.0, 0.0})
	idx.AddItem(2, []float32{0.0, 0.0, 1.0})
	idx.Build(3, 1)

	ctx := idx.CreateContext()
	query := []float32{1.0, 0.1, 0.0}
	resultBefore, _ := idx.GetNnsByVector(query, 3, -1, ctx)

	err := idx.Save(path)
	require.NoError(t, err)

	ctx2 := idx.CreateContext()
	resultAfter, _ := idx.GetNnsByVector(query, 3, -1, ctx2)

	assert.Equal(t, resultBefore, resultAfter,
		"search results should match after Save/Load")

	idx2 := mkIndex()
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	ctx3 := idx2.CreateContext()
	resultFresh, _ := idx2.GetNnsByVector(query, 3, -1, ctx3)

	assert.Equal(t, resultBefore, resultFresh,
		"search results from fresh Load should match original")
}

// TestOneItemSaveLoadSearchUsingBuilderReproducer is based directly on
// the reproduction code from GitHub issue #5.
func TestOneItemSaveLoadSearchUsingBuilderReproducer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")
	dim := 3

	idx := builder.Index().
		AngularDistance(dim).
		UseMultiWorkerPolicy().
		MmapIndexAllocator().
		Build()

	idx.AddItem(0, []float32{0.5, 0.5, 0.0})
	idx.Build(2, -1)

	err := idx.Save(path)
	require.NoError(t, err)
	idx.Close()

	idx2 := builder.Index().
		AngularDistance(dim).
		UseMultiWorkerPolicy().
		MmapIndexAllocator().
		Build()
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	ctx := idx2.CreateContext()
	ids, dists := idx2.GetNnsByVector([]float32{0.5, 0.5, 0.0}, 1, -1, ctx)

	require.Len(t, ids, 1)
	assert.Equal(t, int32(0), ids[0])
	require.Len(t, dists, 1)
	fmt.Println(ids, dists)
}

// TestSmallIndexAllItemsReturned is a regression test for the root dedup
// issue: with dim=10, 5 items, 2 trees, all 5 items must be returned.
func TestSmallIndexAllItemsReturned(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")
	dim := 10
	nItems := 5

	idx := builder.Index().
		AngularDistance(dim).
		SingleWorkerPolicy().
		MmapIndexAllocator().
		Build()
	defer idx.Close()

	vectors := make([][]float32, nItems)
	for i := 0; i < nItems; i++ {
		v := make([]float32, dim)
		v[i%dim] = 1.0
		vectors[i] = v
		require.NoError(t, idx.AddItem(int32(i), v))
	}

	require.NoError(t, idx.Build(2, 1))

	err := idx.Save(path)
	require.NoError(t, err)

	idx2 := builder.Index().
		AngularDistance(dim).
		SingleWorkerPolicy().
		MmapIndexAllocator().
		Build()
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	ctx := idx2.CreateContext()
	// Use large search_k to ensure all items are found
	result, _ := idx2.GetNnsByVector(vectors[0], nItems, nItems*100, ctx)

	require.Len(t, result, nItems,
		"all %d items must be returned after Save/Load", nItems)

	// Verify all item IDs are present
	seen := make(map[int32]bool)
	for _, id := range result {
		seen[id] = true
	}
	for i := 0; i < nItems; i++ {
		assert.True(t, seen[int32(i)], "item %d missing from results", i)
	}
}

func BenchmarkSmallIndexSearch(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "index.ann")

	idx := builder.Index().
		AngularDistance(3).
		SingleWorkerPolicy().
		MmapIndexAllocator().
		Build()
	defer idx.Close()

	for i := 0; i < 5; i++ {
		v := make([]float32, 3)
		v[i%3] = 1.0
		require.NoError(b, idx.AddItem(int32(i), v))
	}
	require.NoError(b, idx.Build(3, 1))

	err := idx.Save(path)
	require.NoError(b, err)

	idx2 := builder.Index().
		AngularDistance(3).
		SingleWorkerPolicy().
		MmapIndexAllocator().
		Build()
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(b, err)

	ctx := idx2.CreateContext()
	query := []float32{1.0, 0.0, 0.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx2.GetNnsByVector(query, 3, -1, ctx)
	}
}

func BenchmarkSmallIndexSearchNoLoad(b *testing.B) {
	idx := builder.Index().
		AngularDistance(3).
		SingleWorkerPolicy().
		Build()
	defer idx.Close()

	for i := 0; i < 5; i++ {
		v := make([]float32, 3)
		v[i%3] = 1.0
		require.NoError(b, idx.AddItem(int32(i), v))
	}
	require.NoError(b, idx.Build(3, 1))

	ctx := idx.CreateContext()
	query := []float32{1.0, 0.0, 0.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.GetNnsByVector(query, 3, -1, ctx)
	}
}
