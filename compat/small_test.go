package compat

import (
	"errors"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/engine"
	"path/filepath"
	"testing"
)

// Preserves the original loaded-small-index panic regression across all metrics.
func TestSmallIndexesAndSparseHoles(t *testing.T) {
	for _, kind := range []codec.Kind{codec.Angular, codec.Dot, codec.Euclidean, codec.Manhattan, codec.Hamming} {
		for _, count := range []int{1, 2, 3, 9, 32} {
			for _, workers := range []int{1, 2} {
				for _, sparse := range []bool{false, true} {
					dim := 3
					if kind == codec.Hamming {
						dim = 130
					}
					b, e := engine.NewBuilder(kind, dim)
					if e != nil {
						t.Fatal(e)
					}
					b.Config.Workers(workers).Trees(3).Seed(42)
					for id := 0; id < count; id++ {
						target := goannoy.ItemID(id)
						if sparse {
							target = goannoy.ItemID(100 + id*7)
						}
						if kind == codec.Hamming {
							e = b.AddWords(target, []uint64{uint64(id), uint64(id * 17), 0})
						} else {
							e = b.AddFloat(target, []float32{float32(id + 1), 1, -1})
						}
						if e != nil {
							t.Fatal(e)
						}
					}
					path := filepath.Join(t.TempDir(), "index.ann")
					i, e := b.BuildTo(nil, path)
					if e != nil {
						t.Fatalf("%s count=%d sparse=%v: %v", kind, count, sparse, e)
					}
					if i.Trees() != 3 {
						t.Fatalf("%s count=%d trees=%d", kind, count, i.Trees())
					}
					s := i.NewSearcher()
					dst := make([]goannoy.Result, count+20)
					for repeat := 0; repeat < 3; repeat++ {
						var n int
						if kind == codec.Hamming {
							n, e = s.SearchWords([]uint64{0, 0, 0}, dst, 10000)
						} else {
							n, e = s.SearchFloat([]float32{1, 1, -1}, dst, 10000)
						}
						if e != nil || n != count {
							t.Fatalf("%s count=%d sparse=%v got=%d %v", kind, count, sparse, n, e)
						}
						seen := map[goannoy.ItemID]bool{}
						for _, r := range dst[:n] {
							if seen[r.ID] {
								t.Fatal("duplicate result")
							}
							seen[r.ID] = true
						}
					}
					if sparse {
						if _, e = s.SearchID(0, dst, -1); !errors.Is(e, goannoy.ErrInvalidID) {
							t.Fatal(e)
						}
					}
					if e := i.Close(); e != nil {
						t.Fatal(e)
					}
				}
			}
		}
	}
}
