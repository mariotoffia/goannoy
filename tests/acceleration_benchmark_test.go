package tests

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/mariotoffia/goannoy/builder"
	"github.com/mariotoffia/goannoy/interfaces"
	gorandom "github.com/mariotoffia/goannoy/random"
)

type accelerationBenchConfig struct {
	distance   string
	dim        int
	items      int
	trees      int
	queryBatch int
}

func BenchmarkAccelerationBuild(b *testing.B) {
	configs := []accelerationBenchConfig{
		{distance: "angular", dim: 128, items: 10_000, trees: 10},
		{distance: "angular", dim: 512, items: 10_000, trees: 10},
		{distance: "dotproduct", dim: 128, items: 10_000, trees: 10},
		{distance: "dotproduct", dim: 512, items: 10_000, trees: 10},
	}

	for _, cfg := range configs {
		cfg := cfg
		b.Run(cfg.name("build"), func(b *testing.B) {
			vectors := benchmarkVectors(cfg.items, cfg.dim, 42)

			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				idx := newAccelerationBenchmarkIndex(cfg.distance, cfg.dim)
				for itemID, v := range vectors {
					if err := idx.AddItem(int32(itemID), v); err != nil {
						b.Fatal(err)
					}
				}

				if err := idx.Build(cfg.trees, 1); err != nil {
					_ = idx.Close()
					b.Fatal(err)
				}

				if err := idx.Close(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkAccelerationSearch(b *testing.B) {
	configs := []accelerationBenchConfig{
		{distance: "angular", dim: 128, items: 10_000, trees: 10, queryBatch: 128},
		{distance: "angular", dim: 512, items: 10_000, trees: 10, queryBatch: 128},
		{distance: "angular", dim: 1536, items: 10_000, trees: 10, queryBatch: 128},
		{distance: "angular", dim: 128, items: 50_000, trees: 10, queryBatch: 128},
		{distance: "dotproduct", dim: 128, items: 10_000, trees: 10, queryBatch: 128},
		{distance: "dotproduct", dim: 512, items: 10_000, trees: 10, queryBatch: 128},
		{distance: "dotproduct", dim: 1536, items: 10_000, trees: 10, queryBatch: 128},
		{distance: "dotproduct", dim: 128, items: 50_000, trees: 10, queryBatch: 128},
	}

	for _, cfg := range configs {
		cfg := cfg
		b.Run(cfg.name("search"), func(b *testing.B) {
			vectors := benchmarkVectors(cfg.items, cfg.dim, 99)
			queries := benchmarkVectors(cfg.queryBatch, cfg.dim, 1234)
			idx := newAccelerationBenchmarkIndex(cfg.distance, cfg.dim)
			b.Cleanup(func() {
				_ = idx.Close()
			})

			for itemID, v := range vectors {
				if err := idx.AddItem(int32(itemID), v); err != nil {
					b.Fatal(err)
				}
			}

			if err := idx.Build(cfg.trees, 1); err != nil {
				b.Fatal(err)
			}

			ctx := idx.CreateContext()
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				query := queries[i%len(queries)]
				idx.GetNnsByVector(query, 10, -1, ctx)
			}
		})
	}
}

func (c accelerationBenchConfig) name(prefix string) string {
	return fmt.Sprintf(
		"%s/distance=%s/dim=%d/items=%d/trees=%d",
		prefix,
		c.distance,
		c.dim,
		c.items,
		c.trees,
	)
}

func benchmarkVectors(items, dim int, seed int64) [][]float32 {
	rng := rand.New(rand.NewSource(seed))
	vectors := make([][]float32, items)

	for i := 0; i < items; i++ {
		v := make([]float32, dim)
		for j := 0; j < dim; j++ {
			v[j] = float32(rng.NormFloat64())
		}
		vectors[i] = v
	}

	return vectors
}

func newAccelerationBenchmarkIndex(distance string, dim int) interfaces.AnnoyIndex[float32] {
	bld := builder.Index().
		Random(gorandom.NewKiss64Random(42)).
		SingleWorkerPolicy()

	switch distance {
	case "angular":
		return bld.AngularDistance(dim).Build()
	case "dotproduct":
		return bld.DotProductDistance(dim).Build()
	default:
		panic("unsupported distance: " + distance)
	}
}
