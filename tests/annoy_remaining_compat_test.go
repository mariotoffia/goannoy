package tests

import (
	"math/rand"
	"os"
	"testing"

	"github.com/mariotoffia/goannoy/builder"
	"github.com/mariotoffia/goannoy/interfaces"
	gorandom "github.com/mariotoffia/goannoy/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnoyGetItemReturnsCopy(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{1, 2, 3}))

	got := idx.GetItem(0)
	got[0] = 99

	assert.Equal(t, []float32{1, 2, 3}, idx.GetItem(0))
}

func TestAnnoyDefaultRandomMatchesKiss64(t *testing.T) {
	makeData := func(seed int64) [][]float32 {
		rng := rand.New(rand.NewSource(seed))
		data := make([][]float32, 256)
		for i := range data {
			v := make([]float32, 8)
			for j := range v {
				v[j] = float32(rng.NormFloat64())
			}
			data[i] = v
		}
		return data
	}

	buildIndex := func(useDefault bool) []int32 {
		var idx interfaces.AnnoyIndex[float32]

		if useDefault {
			idx = angularIdx(8)
		} else {
			idx = builder.Index().
				Random(gorandom.NewKiss64Random(0)).
				AngularDistance(8).
				SingleWorkerPolicy().
				Build()
		}
		defer idx.Close()

		for i, v := range makeData(42) {
			require.NoError(t, idx.AddItem(int32(i), v))
		}
		require.NoError(t, idx.Build(10, 1))

		ctx := idx.CreateContext()
		got, _ := idx.GetNnsByVector([]float32{0.3, -0.5, 0.9, 1.1, -0.2, 0.4, 0.7, -1.3}, 25, -1, ctx)
		return got
	}

	assert.Equal(t, buildIndex(false), buildIndex(true))
}

func TestAnnoyAddAfterBuildReturnsError(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{1, 0, 0}))
	require.NoError(t, idx.Build(2, 1))

	err := idx.AddItem(1, []float32{0, 1, 0})
	require.Error(t, err)
}

func TestAnnoyBuildTwiceReturnsError(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{1, 0, 0}))
	require.NoError(t, idx.Build(2, 1))

	err := idx.Build(2, 1)
	require.Error(t, err)
}

func TestAnnoyEuclideanGetNnsByVector(t *testing.T) {
	idx := builder.Index().
		EuclideanDistance(2).
		SingleWorkerPolicy().
		Build()
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{2, 2}))
	require.NoError(t, idx.AddItem(1, []float32{3, 2}))
	require.NoError(t, idx.AddItem(2, []float32{3, 3}))
	require.NoError(t, idx.Build(10, 1))

	ctx := idx.CreateContext()
	got, _ := idx.GetNnsByVector([]float32{4, 4}, 3, -1, ctx)
	assert.Equal(t, []int32{2, 1, 0}, got)
}

func TestAnnoyManhattanGetNnsByVector(t *testing.T) {
	idx := builder.Index().
		ManhattanDistance(2).
		SingleWorkerPolicy().
		Build()
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{2, 2}))
	require.NoError(t, idx.AddItem(1, []float32{3, 2}))
	require.NoError(t, idx.AddItem(2, []float32{3, 3}))
	require.NoError(t, idx.Build(10, 1))

	ctx := idx.CreateContext()
	got, _ := idx.GetNnsByVector([]float32{5, 3}, 3, -1, ctx)
	assert.Equal(t, []int32{2, 1, 0}, got)
}

func TestAnnoyGetNnsByVectorIDs(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{0, 0, 1}))
	require.NoError(t, idx.AddItem(1, []float32{0, 1, 0}))
	require.NoError(t, idx.AddItem(2, []float32{1, 0, 0}))
	require.NoError(t, idx.Build(10, 1))

	ctx := idx.CreateContext()
	ids := idx.GetNnsByVectorIDs([]float32{3, 2, 1}, 3, -1, ctx)
	assert.Equal(t, []int32{2, 1, 0}, ids)
}

func TestAnnoyGetNItems(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	assert.Equal(t, 0, idx.GetNItems())

	require.NoError(t, idx.AddItem(0, []float32{1, 0, 0}))
	assert.Equal(t, 1, idx.GetNItems())

	require.NoError(t, idx.AddItem(1, []float32{0, 1, 0}))
	assert.Equal(t, 2, idx.GetNItems())

	require.NoError(t, idx.AddItem(2, []float32{0, 0, 1}))
	assert.Equal(t, 3, idx.GetNItems())
}

func TestAnnoyGetNItemsWithGaps(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{1, 0, 0}))
	require.NoError(t, idx.AddItem(5, []float32{0, 1, 0}))
	// GetNItems returns the highest item ID + 1, matching Spotify Annoy behavior.
	assert.Equal(t, 6, idx.GetNItems())
}

func TestAnnoyGetNTrees(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	assert.Equal(t, 0, idx.GetNTrees())

	require.NoError(t, idx.AddItem(0, []float32{1, 0, 0}))
	require.NoError(t, idx.AddItem(1, []float32{0, 1, 0}))
	require.NoError(t, idx.AddItem(2, []float32{0, 0, 1}))
	require.NoError(t, idx.Build(10, 1))

	assert.Equal(t, 10, idx.GetNTrees())
}

func TestAnnoyGetNTreesAfterLoad(t *testing.T) {
	dim := 10
	nItems := 100
	nTrees := 7

	idx := angularIdxMmap(dim)
	defer idx.Close()

	rng := rand.New(rand.NewSource(42))
	for i := 0; i < nItems; i++ {
		v := make([]float32, dim)
		for j := range v {
			v[j] = float32(rng.NormFloat64())
		}
		require.NoError(t, idx.AddItem(int32(i), v))
	}
	require.NoError(t, idx.Build(nTrees, 1))

	tmp := t.TempDir()
	path := tmp + "/test_ntrees.ann"
	require.NoError(t, idx.Save(path))

	// After Save calls Load internally, trees should be preserved.
	assert.Equal(t, nTrees, idx.GetNTrees())
	assert.Equal(t, nItems, idx.GetNItems())
}

func TestAnnoyHammingBasicNns(t *testing.T) {
	idx := builder.Index().
		HammingDistance(6).
		SingleWorkerPolicy().
		Build()
	defer idx.Close()

	u := []float32{1, 0, 1, 0, 1, 0}
	v := []float32{1, 1, 1, 0, 1, 1}

	require.NoError(t, idx.AddItem(0, u))
	require.NoError(t, idx.AddItem(1, v))
	require.NoError(t, idx.Build(10, 1))

	ctx := idx.CreateContext()
	ids, dists := idx.GetNnsByItem(0, 2, -1, ctx)
	assert.Equal(t, []int32{0, 1}, ids)
	require.Len(t, dists, 2)
	assert.Equal(t, float32(0), dists[0])
	assert.Equal(t, float32(2), dists[1])
}

func TestAnnoyUnloadAndReload(t *testing.T) {
	idx := angularIdxMmap(3)
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{0, 0, 1}))
	require.NoError(t, idx.AddItem(1, []float32{0, 1, 0}))
	require.NoError(t, idx.AddItem(2, []float32{1, 0, 0}))
	require.NoError(t, idx.Build(10, 1))

	tmp := t.TempDir()
	path := tmp + "/test_unload.ann"
	require.NoError(t, idx.Save(path))

	assert.Equal(t, 3, idx.GetNItems())

	// Unload releases memory but keeps config.
	require.NoError(t, idx.Unload())
	assert.Equal(t, 0, idx.GetNItems())
	assert.Equal(t, 0, idx.GetNTrees())

	// Reload the same file — should work without error.
	require.NoError(t, idx.Load(path))
	assert.Equal(t, 3, idx.GetNItems())
	assert.Greater(t, idx.GetNTrees(), 0)

	ctx := idx.CreateContext()
	r, _ := idx.GetNnsByVector([]float32{3, 2, 1}, 3, -1, ctx)
	assert.Equal(t, []int32{2, 1, 0}, r)
}

func TestAnnoyPrefaultAfterLoad(t *testing.T) {
	idx := angularIdxMmap(3)
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{0, 0, 1}))
	require.NoError(t, idx.AddItem(1, []float32{0, 1, 0}))
	require.NoError(t, idx.AddItem(2, []float32{1, 0, 0}))
	require.NoError(t, idx.Build(10, 1))

	tmp := t.TempDir()
	path := tmp + "/test_prefault.ann"
	require.NoError(t, idx.Save(path))

	// Prefault should succeed on mmap-backed index.
	require.NoError(t, idx.Prefault())

	// Search should still work after prefault.
	ctx := idx.CreateContext()
	r, _ := idx.GetNnsByVector([]float32{3, 2, 1}, 3, -1, ctx)
	assert.Equal(t, []int32{2, 1, 0}, r)
}

func TestAnnoyPrefaultOnUnloadedIsNoOp(t *testing.T) {
	idx := angularIdxMmap(3)
	defer idx.Close()

	// Prefault on an index with no loaded data should be a no-op.
	require.NoError(t, idx.Prefault())
}

func TestAnnoyUnbuildAndRebuild(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{0, 0, 1}))
	require.NoError(t, idx.AddItem(1, []float32{0, 1, 0}))
	require.NoError(t, idx.Build(5, 1))

	assert.Equal(t, 2, idx.GetNItems())
	assert.Equal(t, 5, idx.GetNTrees())

	// Unbuild removes trees but preserves items.
	require.NoError(t, idx.Unbuild())
	assert.Equal(t, 2, idx.GetNItems())
	assert.Equal(t, 0, idx.GetNTrees())

	// Add more items and rebuild.
	require.NoError(t, idx.AddItem(2, []float32{1, 0, 0}))
	assert.Equal(t, 3, idx.GetNItems())

	require.NoError(t, idx.Build(8, 1))
	assert.Equal(t, 8, idx.GetNTrees())

	ctx := idx.CreateContext()
	r, _ := idx.GetNnsByVector([]float32{3, 2, 1}, 3, -1, ctx)
	assert.Equal(t, []int32{2, 1, 0}, r)
}

func TestAnnoyUnbuildOnUnbuiltReturnsError(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{1, 0, 0}))
	require.Error(t, idx.Unbuild())
}

func TestAnnoyUnbuildOnLoadedReturnsError(t *testing.T) {
	idx := angularIdxMmap(3)
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{0, 0, 1}))
	require.NoError(t, idx.AddItem(1, []float32{0, 1, 0}))
	require.NoError(t, idx.Build(5, 1))

	tmp := t.TempDir()
	path := tmp + "/test_unbuild_loaded.ann"
	require.NoError(t, idx.Save(path))

	// After Save→Load, the index is "loaded" — unbuild should fail.
	require.Error(t, idx.Unbuild())
}

func TestAnnoyOnDiskBuild(t *testing.T) {
	idx := angularIdxMmap(3)
	defer idx.Close()

	tmp := t.TempDir()
	path := tmp + "/ondisk.ann"
	require.NoError(t, idx.OnDiskBuild(path))

	require.NoError(t, idx.AddItem(0, []float32{0, 0, 1}))
	require.NoError(t, idx.AddItem(1, []float32{0, 1, 0}))
	require.NoError(t, idx.AddItem(2, []float32{1, 0, 0}))
	require.NoError(t, idx.Build(10, 1))

	// After on-disk build, the index is loaded and searchable.
	assert.Equal(t, 3, idx.GetNItems())
	assert.Greater(t, idx.GetNTrees(), 0)

	ctx := idx.CreateContext()
	r, _ := idx.GetNnsByVector([]float32{3, 2, 1}, 3, -1, ctx)
	assert.Equal(t, []int32{2, 1, 0}, r)

	// The file should exist and be loadable by a fresh index.
	_, err := os.Stat(path)
	require.NoError(t, err)
}

func TestAnnoyOnDiskBuildLoadBySeparateIndex(t *testing.T) {
	idx := angularIdxMmap(3)
	defer idx.Close()

	tmp := t.TempDir()
	path := tmp + "/ondisk2.ann"
	require.NoError(t, idx.OnDiskBuild(path))

	require.NoError(t, idx.AddItem(0, []float32{0, 0, 1}))
	require.NoError(t, idx.AddItem(1, []float32{0, 1, 0}))
	require.NoError(t, idx.AddItem(2, []float32{1, 0, 0}))
	require.NoError(t, idx.Build(10, 1))

	// Load the on-disk-built file using a completely new index.
	idx2 := angularIdxMmap(3)
	defer idx2.Close()

	require.NoError(t, idx2.Load(path))
	assert.Equal(t, 3, idx2.GetNItems())

	ctx := idx2.CreateContext()
	r, _ := idx2.GetNnsByVector([]float32{3, 2, 1}, 3, -1, ctx)
	assert.Equal(t, []int32{2, 1, 0}, r)
}

func TestAnnoyOnDiskBuildAfterAddItemReturnsError(t *testing.T) {
	idx := angularIdxMmap(3)
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{1, 0, 0}))
	require.Error(t, idx.OnDiskBuild("/tmp/bad.ann"))
}

func TestAnnoyOnDiskBuildLargerIndex(t *testing.T) {
	dim := 10
	nItems := 500
	nTrees := 5

	idx := angularIdxMmap(dim)
	defer idx.Close()

	tmp := t.TempDir()
	path := tmp + "/ondisk_large.ann"
	require.NoError(t, idx.OnDiskBuild(path))

	rng := rand.New(rand.NewSource(42))
	vectors := make([][]float32, nItems)
	for i := range nItems {
		v := make([]float32, dim)
		for j := range v {
			v[j] = float32(rng.NormFloat64())
		}
		vectors[i] = v
		require.NoError(t, idx.AddItem(int32(i), v))
	}
	require.NoError(t, idx.Build(nTrees, 1))

	assert.Equal(t, nItems, idx.GetNItems())
	assert.Equal(t, nTrees, idx.GetNTrees())

	// Verify self-search: each item should find itself as nearest neighbor.
	ctx := idx.CreateContext()
	selfFound := 0
	for i := range nItems {
		ids := idx.GetNnsByItemIDs(int32(i), 1, nItems*10, ctx)
		if len(ids) > 0 && ids[0] == int32(i) {
			selfFound++
		}
	}
	precision := float64(selfFound) / float64(nItems)
	assert.Greater(t, precision, 0.95, "self-search precision should be high")
}
