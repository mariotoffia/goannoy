package engine

import (
	"context"
	"errors"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"math"
	"testing"
)

func TestZeroCoincidentAndDenormalVectors(t *testing.T) {
	for _, kind := range []codec.Kind{codec.Angular, codec.Dot, codec.Euclidean, codec.Manhattan, codec.Hamming} {
		for _, value := range []float32{0, math.SmallestNonzeroFloat32, 1} {
			b, e := NewBuilder(kind, 1)
			if e != nil {
				t.Fatal(e)
			}
			b.Config.Workers(1).Trees(3)
			for id := 0; id < 64; id++ {
				if kind == codec.Hamming {
					e = b.AddWords(goannoy.ItemID(id), []uint64{0})
				} else {
					e = b.AddFloat(goannoy.ItemID(id), []float32{value})
				}
				if e != nil {
					t.Fatal(e)
				}
			}
			i, e := b.Build(nil)
			if e != nil {
				t.Fatalf("%s value=%g: %v", kind, value, e)
			}
			s := i.NewSearcher()
			dst := make([]goannoy.Result, 64)
			var n int
			if kind == codec.Hamming {
				n, e = s.SearchWords([]uint64{0}, dst, 100000)
			} else {
				n, e = s.SearchFloat([]float32{value}, dst, 100000)
			}
			if e != nil || n != 64 {
				t.Fatalf("%s: %d %v", kind, n, e)
			}
			for id, r := range dst {
				if int(r.ID) != id || math.IsNaN(float64(r.Distance)) {
					t.Fatalf("%s: %+v", kind, r)
				}
			}
			if e := i.Close(); e != nil {
				t.Fatal(e)
			}
		}
	}
}

func TestCancellationMidBuildPreservesInput(t *testing.T) {
	b, _ := NewBuilder(codec.Angular, 3)
	b.Config.Trees(1000).Workers(2)
	for id := 0; id < 256; id++ {
		if e := b.AddFloat(goannoy.ItemID(id), fixtureFloat(id)); e != nil {
			t.Fatal(e)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, e := b.Build(ctx); done <- e }()
	cancel()
	if e := <-done; !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	b.Config.Trees(1).Workers(1)
	i, e := b.Build(nil)
	if e != nil {
		t.Fatal(e)
	}
	defer i.Close()
}
