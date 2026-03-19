package tests

import (
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// BenchmarkBuildSaveLoadSearch measures end-to-end performance for a
// medium-sized index: build, save, load, and search.
func BenchmarkBuildSaveLoadSearch(b *testing.B) {
	dim := 10
	nItems := 1000
	nTrees := 5

	rng := rand.New(rand.NewSource(42))
	vectors := make([][]float32, nItems)
	for i := 0; i < nItems; i++ {
		v := make([]float32, dim)
		for d := 0; d < dim; d++ {
			v[d] = rng.Float32()*2 - 1
		}
		vectors[i] = v
	}

	query := vectors[0]
	dir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := filepath.Join(dir, "bench.ann")

		idx := angularIdxMmap(dim)
		for j := 0; j < nItems; j++ {
			require.NoError(b, idx.AddItem(int32(j), vectors[j]))
		}
		require.NoError(b, idx.Build(nTrees, 1))

		err := idx.Save(path)
		if err != nil {
			b.Fatal(err)
		}

		idx2 := angularIdxMmap(dim)
		err = idx2.Load(path)
		if err != nil {
			b.Fatal(err)
		}

		ctx := idx2.CreateContext()
		idx2.GetNnsByVector(query, 10, -1, ctx)

		_ = idx2.Close()
		_ = idx.Close()
	}
}

// BenchmarkSearchAfterLoad measures search-only performance on a loaded index.
func BenchmarkSearchAfterLoad(b *testing.B) {
	dim := 10
	nItems := 1000
	nTrees := 5

	dir := b.TempDir()
	path := filepath.Join(dir, "bench.ann")

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
		require.NoError(b, idx.AddItem(int32(i), v))
	}

	require.NoError(b, idx.Build(nTrees, 1))

	err := idx.Save(path)
	require.NoError(b, err)

	idx2 := angularIdxMmap(dim)
	defer func() { _ = idx2.Close() }()

	err = idx2.Load(path)
	require.NoError(b, err)

	ctx := idx2.CreateContext()
	query := vectors[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx2.GetNnsByVector(query, 10, -1, ctx)
	}
}

// BenchmarkMultiDimSearch benchmarks search across different dimensions to
// expose dimension-dependent performance characteristics.
func BenchmarkMultiDimSearch(b *testing.B) {
	dims := []int{3, 10, 50, 100}
	nItems := 500
	nTrees := 3

	for _, dim := range dims {
		b.Run(dimName(dim), func(b *testing.B) {
			dir := b.TempDir()
			path := filepath.Join(dir, "bench.ann")

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
				require.NoError(b, idx.AddItem(int32(i), v))
			}

			require.NoError(b, idx.Build(nTrees, 1))

			err := idx.Save(path)
			require.NoError(b, err)

			idx2 := angularIdxMmap(dim)
			defer idx2.Close()

			err = idx2.Load(path)
			require.NoError(b, err)

			ctx := idx2.CreateContext()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				idx2.GetNnsByVector(query, 10, -1, ctx)
			}
		})
	}
}

func dimName(dim int) string {
	switch dim {
	case 3:
		return "dim3"
	case 10:
		return "dim10"
	case 50:
		return "dim50"
	case 100:
		return "dim100"
	default:
		return "dimN"
	}
}
