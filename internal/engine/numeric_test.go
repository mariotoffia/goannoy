package engine

import (
	"errors"
	"testing"

	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
)

func TestFiniteVectorOverflowReturnsError(t *testing.T) {
	for _, kind := range []codec.Kind{codec.Angular, codec.Dot, codec.Euclidean, codec.Manhattan} {
		t.Run(kind.String(), func(t *testing.T) {
			b, _ := NewBuilder(kind, 1)
			b.Config.Workers(1).Trees(1)
			if err := b.AddFloat(0, []float32{1e10}); err != nil {
				t.Fatal(err)
			}
			i, err := b.Build(nil)
			if err != nil {
				t.Fatal(err)
			}
			defer i.Close()
			if err := b.AddFloat(0, []float32{1e30}); !errors.Is(err, goannoy.ErrInvalidVector) {
				t.Errorf("Add overflow: %v", err)
			}
			if _, err := i.NewSearcher().SearchFloat([]float32{1e30}, make([]goannoy.Result, 1), -1); !errors.Is(err, goannoy.ErrInvalidVector) {
				t.Errorf("query overflow: %v", err)
			}
			// Rejected replacement must not mutate the builder's previous item.
			next, err := b.Build(nil)
			if err != nil {
				t.Fatal(err)
			}
			dst := make([]float32, 1)
			if err := next.VectorFloat(0, dst); err != nil || dst[0] != 1e10 {
				t.Errorf("failed Add changed item: %v %v", dst, err)
			}
			next.Close()
		})
	}
}

func TestDerivedDistanceOverflowReturnsError(t *testing.T) {
	b, _ := NewBuilder(codec.Euclidean, 1)
	b.Config.Workers(1).Trees(1)
	for id, v := range []float32{1e19, -1e19} {
		if err := b.AddFloat(goannoy.ItemID(id), []float32{v}); err != nil {
			t.Fatal(err)
		}
	}
	i, err := b.Build(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer i.Close()
	if _, err := i.NewSearcher().SearchFloat([]float32{-1e19}, make([]goannoy.Result, 2), 100); !errors.Is(err, goannoy.ErrInvalidVector) {
		t.Errorf("search overflow: %v", err)
	}
	if _, err := i.Distance(0, 1); !errors.Is(err, goannoy.ErrInvalidVector) {
		t.Errorf("distance overflow: %v", err)
	}
}
