package tests

import (
	"math/rand"
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
