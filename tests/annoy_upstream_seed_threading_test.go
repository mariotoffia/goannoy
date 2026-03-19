package tests

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ported or adapted from Spotify Annoy test/seed_test.py,
// test/threading_test.py, and test/multithreaded_build_test.py.

func TestUpstreamSeedDeterminism(t *testing.T) {
	const (
		f        = 10
		nItems   = 1000
		nQueries = 50
	)

	rng := rand.New(rand.NewSource(42))
	data := make([][]float32, nItems)
	queries := make([][]float32, nQueries)

	for i := range data {
		v := make([]float32, f)
		for z := range v {
			v[z] = rng.Float32()
		}
		data[i] = v
	}

	for i := range queries {
		v := make([]float32, f)
		for z := range v {
			v[z] = rng.Float32()
		}
		queries[i] = v
	}

	indexes := make([]interfaces.AnnoyIndex[float32], 0, 2)

	for i := 0; i < 2; i++ {
		idx := angularIdxWithSeed(f, 42)
		for j, vec := range data {
			require.NoError(t, idx.AddItem(int32(j), vec))
		}
		require.NoError(t, idx.Build(10, -1))
		indexes = append(indexes, idx)
		defer idx.Close()
	}

	for _, q := range queries {
		ctx0 := indexes[0].CreateContext()
		ctx1 := indexes[1].CreateContext()
		got0, _ := indexes[0].GetNnsByVector(q, 100, -1, ctx0)
		got1, _ := indexes[1].GetNnsByVector(q, 100, -1, ctx1)
		assert.Equal(t, got0, got1)
	}
}

func TestUpstreamConcurrentQueries(t *testing.T) {
	const (
		n = 5000
		f = 10
	)

	idx := angularIdxMultiWithSeed(f, 123)
	defer idx.Close()

	rng := rand.New(rand.NewSource(123))
	for j := 0; j < n; j++ {
		v := make([]float32, f)
		for z := range v {
			v[z] = float32(rng.NormFloat64())
		}
		require.NoError(t, idx.AddItem(int32(j), v))
	}
	require.NoError(t, idx.Build(10, 4))

	var wg sync.WaitGroup
	errCh := make(chan error, 256)

	for j := 0; j < 256; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx := idx.CreateContext()
			ids, _ := idx.GetNnsByItem(1, 100, -1, ctx)
			if len(ids) == 0 {
				errCh <- assert.AnError
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}
}

func TestUpstreamBuildingWithThreads(t *testing.T) {
	for _, workers := range []int{1, 2, 4, 8} {
		t.Run(fmt.Sprintf("workers=%d", workers), func(t *testing.T) {
			const (
				n      = 1000
				f      = 10
				nTrees = 31
			)

			idx := angularIdxMultiWithSeed(f, 77)
			defer idx.Close()

			rng := rand.New(rand.NewSource(int64(workers)))
			for j := 0; j < n; j++ {
				v := make([]float32, f)
				for z := range v {
					v[z] = float32(rng.NormFloat64())
				}
				require.NoError(t, idx.AddItem(int32(j), v))
			}

			require.NoError(t, idx.Build(nTrees, workers))

			ctx := idx.CreateContext()
			ids, dists := idx.GetNnsByItem(0, 10, -1, ctx)
			require.NotEmpty(t, ids)
			require.NotEmpty(t, dists)
		})
	}
}
