package tests

import (
	"math/rand"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ported or adapted from Spotify Annoy test/index_test.py.

func TestUpstreamNotFoundTree(t *testing.T) {
	idx := angularIdxMmap(10)
	defer idx.Close()

	err := idx.Load(filepath.Join(t.TempDir(), "nonexists.tree"))
	require.Error(t, err)
}

func TestUpstreamBinaryCompatibility(t *testing.T) {
	idx := angularIdxMmap(10)
	defer idx.Close()

	err := idx.Load(upstreamBinaryFixturePath())
	require.NoError(t, err)

	ctx := idx.CreateContext()
	result, _ := idx.GetNnsByItem(0, 10, -1, ctx)

	assert.Equal(t, []int32{0, 85, 42, 11, 54, 38, 53, 66, 19, 31}, result)
}

func TestUpstreamLoadSaveFixture(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fixture-copy.tree")

	idx := angularIdxMmap(10)
	defer idx.Close()

	err := idx.Load(upstreamBinaryFixturePath())
	require.NoError(t, err)

	before := idx.GetItem(99)
	err = idx.Save(path)
	require.NoError(t, err)

	afterSave := idx.GetItem(99)
	assert.Equal(t, before, afterSave)

	reloaded := angularIdxMmap(10)
	defer reloaded.Close()

	err = reloaded.Load(path)
	require.NoError(t, err)

	afterReload := reloaded.GetItem(99)
	assert.Equal(t, before, afterReload)
}

func TestUpstreamFailSave(t *testing.T) {
	idx := angularIdxMmap(3)
	defer idx.Close()

	idx.AddItem(0, []float32{1, 0, 0})
	idx.Build(1, 1)

	err := idx.Save("")
	require.Error(t, err)
}

func TestUpstreamOverwriteIndex(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("overwriting a mapped file is not portable on Windows")
	}

	const f = 40
	path := filepath.Join(t.TempDir(), "test.ann")
	rng := rand.New(rand.NewSource(42))

	t1 := angularIdxMmap(f)
	defer t1.Close()
	for i := 0; i < 1000; i++ {
		v := make([]float32, f)
		for z := range v {
			v[z] = float32(rng.NormFloat64())
		}
		require.NoError(t, t1.AddItem(int32(i), v))
	}
	require.NoError(t, t1.Build(10, -1))
	require.NoError(t, t1.Save(path))

	t2 := angularIdxMmap(f)
	defer t2.Close()
	require.NoError(t, t2.Load(path))

	t3 := angularIdxMmap(f)
	defer t3.Close()
	for i := 0; i < 500; i++ {
		v := make([]float32, f)
		for z := range v {
			v[z] = float32(rng.NormFloat64())
		}
		require.NoError(t, t3.AddItem(int32(i), v))
	}
	require.NoError(t, t3.Build(10, -1))
	require.NoError(t, t3.Save(path))

	ctx := t2.CreateContext()
	query := make([]float32, f)
	for z := range query {
		query[z] = float32(rng.NormFloat64())
	}

	assert.NotPanics(t, func() {
		_, _ = t2.GetNnsByVector(query, 100, -1, ctx)
	})
}

func TestUpstreamWriteFailed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "x", "y", "z.annoy")

	idx := angularIdxMmap(40)
	defer idx.Close()

	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 1000; i++ {
		v := make([]float32, 40)
		for z := range v {
			v[z] = float32(rng.NormFloat64())
		}
		require.NoError(t, idx.AddItem(int32(i), v))
	}
	require.NoError(t, idx.Build(10, -1))

	err := idx.Save(path)
	require.Error(t, err)
}

func TestUpstreamDimensionMismatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.annoy")
	rng := rand.New(rand.NewSource(99))

	t1 := angularIdxMmap(100)
	defer t1.Close()
	for i := 0; i < 1000; i++ {
		v := make([]float32, 100)
		for z := range v {
			v[z] = float32(rng.NormFloat64())
		}
		require.NoError(t, t1.AddItem(int32(i), v))
	}
	require.NoError(t, t1.Build(10, -1))
	require.NoError(t, t1.Save(path))

	t2 := angularIdxMmap(200)
	defer t2.Close()
	err := t2.Load(path)
	require.Error(t, err)

	t3 := angularIdxMmap(50)
	defer t3.Close()
	err = t3.Load(path)
	require.Error(t, err)
}

func TestUpstreamAddAfterSave(t *testing.T) {
	idx := angularIdxMmap(100)
	defer idx.Close()

	rng := rand.New(rand.NewSource(88))
	for i := 0; i < 1000; i++ {
		v := make([]float32, 100)
		for z := range v {
			v[z] = float32(rng.NormFloat64())
		}
		require.NoError(t, idx.AddItem(int32(i), v))
	}
	require.NoError(t, idx.Build(10, -1))
	require.NoError(t, idx.Save(filepath.Join(t.TempDir(), "test.annoy")))

	v := make([]float32, 100)
	for z := range v {
		v[z] = float32(rng.NormFloat64())
	}

	require.Error(t, idx.AddItem(1001, v))
}
