package tests

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ported from Spotify Annoy test/holes_test.py.

func TestUpstreamRandomHoles(t *testing.T) {
	const (
		f          = 10
		totalSlots = 2000
		numItems   = 1000
	)

	idx := angularIdx(f)
	defer idx.Close()

	rng := rand.New(rand.NewSource(42))
	validIndices := rng.Perm(totalSlots)[:numItems]
	validSet := make(map[int32]struct{}, len(validIndices))

	for _, i := range validIndices {
		v := make([]float32, f)
		for z := range v {
			v[z] = float32(rng.NormFloat64())
		}

		id := int32(i)
		validSet[id] = struct{}{}
		require.NoError(t, idx.AddItem(id, v))
	}

	require.NoError(t, idx.Build(10, -1))
	ctx := idx.CreateContext()

	for _, i := range validIndices {
		js, _ := idx.GetNnsByItem(int32(i), 10000, -1, ctx)
		for _, j := range js {
			_, ok := validSet[j]
			assert.True(t, ok, "unexpected hole id %d returned for item %d", j, i)
		}
	}

	for i := 0; i < 1000; i++ {
		v := make([]float32, f)
		for z := range v {
			v[z] = float32(rng.NormFloat64())
		}

		js, _ := idx.GetNnsByVector(v, 10000, -1, ctx)
		for _, j := range js {
			_, ok := validSet[j]
			assert.True(t, ok, "unexpected hole id %d returned for query %d", j, i)
		}
	}
}

func testUpstreamHolesBase(t *testing.T, n int) {
	t.Helper()

	const (
		f     = 100
		base  = 100000
		trees = 100
	)

	idx := angularIdx(f)
	defer idx.Close()

	rng := rand.New(rand.NewSource(int64(1000 + n)))
	expected := make(map[int32]struct{}, n)

	for i := 0; i < n; i++ {
		v := make([]float32, f)
		for z := range v {
			v[z] = float32(rng.NormFloat64())
		}

		id := int32(base + i)
		expected[id] = struct{}{}
		require.NoError(t, idx.AddItem(id, v))
	}

	require.NoError(t, idx.Build(trees, -1))
	ctx := idx.CreateContext()
	res, _ := idx.GetNnsByItem(int32(base), n, -1, ctx)
	require.Len(t, res, n)

	actual := make(map[int32]struct{}, len(res))
	for _, id := range res {
		actual[id] = struct{}{}
	}

	assert.Equal(t, expected, actual)
}

func TestUpstreamRootOneChildWithHoles(t *testing.T) {
	testUpstreamHolesBase(t, 1)
}

func TestUpstreamRootTwoChildrenWithHoles(t *testing.T) {
	testUpstreamHolesBase(t, 2)
}

func TestUpstreamRootSomeChildrenWithHoles(t *testing.T) {
	testUpstreamHolesBase(t, 10)
}

func TestUpstreamRootManyChildrenWithHoles(t *testing.T) {
	testUpstreamHolesBase(t, 1000)
}
