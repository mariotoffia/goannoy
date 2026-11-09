package compat

import (
	"errors"
	"math"
	"testing"

	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/angular"
	"github.com/mariotoffia/goannoy/dot"
	"github.com/mariotoffia/goannoy/euclidean"
	"github.com/mariotoffia/goannoy/hamming"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/engine"
	"github.com/mariotoffia/goannoy/manhattan"
)

func checkSearchLimit[B floatBuilder[B, I], I floatIndex[S], S floatSearcher](t *testing.T, constructor func(int) (B, error)) {
	b, err := constructor(1)
	if err != nil {
		t.Fatal(err)
	}
	b.Trees(1).Workers(1)
	if err := b.Add(0, []float32{1}); err != nil {
		t.Fatal(err)
	}
	i, err := b.Build(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer i.Close()
	if _, err := i.NewSearcher().Search([]float32{1}, math.MaxInt, -1); !errors.Is(err, goannoy.ErrInvalidOption) {
		t.Fatal(err)
	}
	if r, err := i.NewSearcher().Search([]float32{1}, 1000, -1); err != nil || len(r) != 1 {
		t.Fatalf("%v %v", r, err)
	}
}
func TestAllocatingSearchRejectsOverflow(t *testing.T) {
	t.Run("angular", func(t *testing.T) {
		checkSearchLimit[*angular.Builder, *angular.Index, *angular.Searcher](t, angular.NewBuilder)
	})
	t.Run("dot", func(t *testing.T) { checkSearchLimit[*dot.Builder, *dot.Index, *dot.Searcher](t, dot.NewBuilder) })
	t.Run("euclidean", func(t *testing.T) {
		checkSearchLimit[*euclidean.Builder, *euclidean.Index, *euclidean.Searcher](t, euclidean.NewBuilder)
	})
	t.Run("manhattan", func(t *testing.T) {
		checkSearchLimit[*manhattan.Builder, *manhattan.Index, *manhattan.Searcher](t, manhattan.NewBuilder)
	})
	t.Run("hamming", func(t *testing.T) {
		b, _ := hamming.NewBuilder(1)
		b.Trees(1).Workers(1)
		if err := b.Add(0, []uint64{1}); err != nil {
			t.Fatal(err)
		}
		i, err := b.Build(nil)
		if err != nil {
			t.Fatal(err)
		}
		defer i.Close()
		if _, err := i.NewSearcher().Search([]uint64{1}, math.MaxInt, -1); !errors.Is(err, goannoy.ErrInvalidOption) {
			t.Fatal(err)
		}
	})
}
func TestHammingRejectsInexactDimension(t *testing.T) {
	for _, dim := range []int{1<<24 + 1, 1<<24 + 2} {
		if _, err := hamming.NewBuilder(dim); !errors.Is(err, goannoy.ErrInvalidDimension) {
			t.Errorf("builder dimension %d: %v", dim, err)
		}
		if _, err := hamming.Open("fixtures/hamming.ann", dim); !errors.Is(err, goannoy.ErrInvalidDimension) {
			t.Errorf("open dimension %d: %v", dim, err)
		}
	}
	if _, err := hamming.NewBuilder(1 << 24); err != nil {
		t.Fatal(err)
	}
}
func TestCompactSparseRoots(t *testing.T) {
	for _, kind := range []codec.Kind{codec.Angular, codec.Dot, codec.Euclidean, codec.Manhattan, codec.Hamming} {
		for _, id := range []goannoy.ItemID{1, 2, 3, 4} {
			dim := 3
			if kind == codec.Hamming {
				dim = 130
			}
			b, _ := engine.NewBuilder(kind, dim)
			b.Config.Trees(3).Workers(1)
			var err error
			if kind == codec.Hamming {
				err = b.AddWords(id, []uint64{1, 0, 0})
			} else {
				err = b.AddFloat(id, []float32{1, 0, 0})
			}
			if err != nil {
				t.Fatal(err)
			}
			i, err := b.Build(nil)
			if err != nil {
				t.Fatalf("%s id%d: %v", kind, id, err)
			}
			dst := make([]goannoy.Result, 10)
			n, err := i.NewSearcher().SearchID(id, dst, 1000)
			if err != nil || n != 1 || dst[0].ID != id {
				t.Fatalf("%s id%d: %v n%d %v", kind, id, dst, n, err)
			}
			i.Close()
		}
	}
}
