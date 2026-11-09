package compat

import (
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/engine"
	"math"
	"math/bits"
	"math/rand"
	"slices"
	"testing"
)

func TestAllMetricsMatchIndependentExactNeighbors(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	const count = 256
	vectors := make([][]float32, count)
	words := make([][]uint64, count)
	for id := range vectors {
		vectors[id] = make([]float32, 8)
		for j := range vectors[id] {
			vectors[id][j] = float32(rng.NormFloat64())
		}
		words[id] = []uint64{rng.Uint64(), rng.Uint64(), rng.Uint64() & 3}
	}
	for _, kind := range []codec.Kind{codec.Angular, codec.Dot, codec.Euclidean, codec.Manhattan, codec.Hamming} {
		t.Run(kind.String(), func(t *testing.T) {
			dim := 8
			if kind == codec.Hamming {
				dim = 130
			}
			b, e := engine.NewBuilder(kind, dim)
			if e != nil {
				t.Fatal(e)
			}
			b.Config.Trees(5).Workers(2).Seed(42)
			for id := range vectors {
				if kind == codec.Hamming {
					e = b.AddWords(goannoy.ItemID(id), words[id])
				} else {
					e = b.AddFloat(goannoy.ItemID(id), vectors[id])
				}
				if e != nil {
					t.Fatal(e)
				}
			}
			i, e := b.Build(nil)
			if e != nil {
				t.Fatal(e)
			}
			defer i.Close()
			s := i.NewSearcher()
			for query := 0; query < 8; query++ {
				q := vectors[query*13]
				qw := words[query*13]
				want := make([]goannoy.Result, count)
				for id, v := range vectors {
					var value float64
					if kind == codec.Hamming {
						for j, x := range qw {
							value += float64(bits.OnesCount64(x ^ words[id][j]))
						}
					} else {
						var aa, bb, ab float64
						for j, x := range q {
							a, b := float64(x), float64(v[j])
							aa += a * a
							bb += b * b
							ab += a * b
							if kind == codec.Euclidean {
								value += (a - b) * (a - b)
							}
							if kind == codec.Manhattan {
								value += math.Abs(a - b)
							}
						}
						if kind == codec.Angular {
							value = math.Sqrt(max(0, 2-2*ab/math.Sqrt(aa*bb)))
						}
						if kind == codec.Dot {
							value = -ab
						}
						if kind == codec.Euclidean {
							value = math.Sqrt(value)
						}
					}
					want[id] = goannoy.Result{ID: goannoy.ItemID(id), Distance: float32(value)}
				}
				slices.SortFunc(want, func(a, b goannoy.Result) int {
					if a.Distance < b.Distance || a.Distance == b.Distance && a.ID < b.ID {
						return -1
					}
					if a.ID == b.ID {
						return 0
					}
					return 1
				})
				got := make([]goannoy.Result, 10)
				var n int
				if kind == codec.Hamming {
					n, e = s.SearchWords(qw, got, 100000)
				} else {
					n, e = s.SearchFloat(q, got, 100000)
				}
				if e != nil || n != 10 {
					t.Fatalf("%d %v", n, e)
				}
				for j, r := range got {
					if r.ID != want[j].ID {
						t.Fatalf("query %d rank %d got %v want %v", query, j, r, want[j])
					}
					d := want[j].Distance
					if kind == codec.Dot {
						d = -d
					}
					if math.Abs(float64(r.Distance-d)) > 1e-3 {
						t.Fatalf("distance %g want %g", r.Distance, d)
					}
				}
			}
		})
	}
}
