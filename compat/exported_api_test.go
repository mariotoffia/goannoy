package compat

import (
	"context"
	"errors"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/angular"
	"github.com/mariotoffia/goannoy/dot"
	"github.com/mariotoffia/goannoy/euclidean"
	"github.com/mariotoffia/goannoy/manhattan"
	"path/filepath"
	"testing"
)

type floatBuilder[B, I any] interface {
	Trees(int) B
	Workers(int) B
	Seed(uint64) B
	Capacity(int) B
	Profile(goannoy.Profile) B
	OnDisk(string) B
	Err() error
	Add(goannoy.ItemID, []float32) error
	Build(context.Context) (I, error)
	BuildTo(context.Context, string) (I, error)
}
type floatIndex[S any] interface {
	Dimension() int
	Items() int
	Trees() int
	Close() error
	NewSearcher() S
	VectorInto(goannoy.ItemID, []float32) error
	Vector(goannoy.ItemID) ([]float32, error)
	Distance(goannoy.ItemID, goannoy.ItemID) (float32, error)
}
type floatSearcher interface {
	SearchInto([]float32, []goannoy.Result, goannoy.SearchK) (int, error)
	Search([]float32, int, goannoy.SearchK) ([]goannoy.Result, error)
	SearchByIDInto(goannoy.ItemID, []goannoy.Result, goannoy.SearchK) (int, error)
}

func exerciseFloat[B floatBuilder[B, I], I floatIndex[S], S floatSearcher](t *testing.T, newBuilder func(int) (B, error)) {
	t.Helper()
	if _, e := newBuilder(0); !errors.Is(e, goannoy.ErrInvalidDimension) {
		t.Fatal(e)
	}
	b, e := newBuilder(2)
	if e != nil {
		t.Fatal(e)
	}
	b.Trees(3).Workers(1).Seed(42).Capacity(32).Profile(goannoy.Throughput)
	if e := b.Err(); e != nil {
		t.Fatal(e)
	}
	for id := 0; id < 32; id++ {
		if e := b.Add(goannoy.ItemID(id), []float32{float32(32 - id), 1}); e != nil {
			t.Fatal(e)
		}
	}
	i, e := b.Build(nil)
	if e != nil {
		t.Fatal(e)
	}
	if i.Dimension() != 2 || i.Items() != 32 || i.Trees() != 3 {
		t.Fatal("metadata")
	}
	q := []float32{32, 1}
	dst := make([]goannoy.Result, 10)
	s := i.NewSearcher()
	if n, e := s.SearchInto(q, dst, 10000); e != nil || n != 10 || dst[0].ID != 0 {
		t.Fatalf("%v n=%d %v", dst, n, e)
	}
	if n, e := s.SearchByIDInto(0, dst, 10000); e != nil || n != 10 {
		t.Fatalf("%d %v", n, e)
	}
	if _, e := s.SearchByIDInto(-1, dst, -1); !errors.Is(e, goannoy.ErrInvalidID) {
		t.Fatal(e)
	}
	if r, e := s.Search(q, 10, 10000); e != nil || len(r) != 10 {
		t.Fatalf("%v %v", r, e)
	}
	if _, e := s.Search(q, -1, -1); !errors.Is(e, goannoy.ErrInvalidOption) {
		t.Fatal(e)
	}
	if _, e := s.Search(nil, 10, -1); !errors.Is(e, goannoy.ErrInvalidVector) {
		t.Fatal(e)
	}
	v, e := i.Vector(0)
	if e != nil || v[0] != 32 {
		t.Fatalf("%v %v", v, e)
	}
	if e := i.VectorInto(0, v); e != nil {
		t.Fatal(e)
	}
	if e := i.VectorInto(0, []float32{1}); !errors.Is(e, goannoy.ErrInvalidVector) {
		t.Fatal(e)
	}
	if _, e := i.Vector(-1); !errors.Is(e, goannoy.ErrInvalidID) {
		t.Fatal(e)
	}
	if _, e := i.Distance(0, 1); e != nil {
		t.Fatal(e)
	}
	if _, e := i.Distance(-1, 0); !errors.Is(e, goannoy.ErrInvalidID) {
		t.Fatal(e)
	}
	if e := i.Close(); e != nil {
		t.Fatal(e)
	}
	if e := i.VectorInto(0, v); !errors.Is(e, goannoy.ErrClosed) {
		t.Fatal(e)
	}
	if _, e := i.Distance(0, 1); !errors.Is(e, goannoy.ErrClosed) {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "index.ann")
	j, e := b.BuildTo(nil, path)
	if e != nil {
		t.Fatal(e)
	}
	if e := j.Close(); e != nil {
		t.Fatal(e)
	}
	b.OnDisk(path)
	j, e = b.Build(nil)
	if e != nil {
		t.Fatal(e)
	}
	if e := j.Close(); e != nil {
		t.Fatal(e)
	}
	b.Trees(0)
	if !errors.Is(b.Err(), goannoy.ErrInvalidOption) {
		t.Fatal(b.Err())
	}
	if e := b.Add(0, q); !errors.Is(e, goannoy.ErrInvalidOption) {
		t.Fatal(e)
	}
	if _, e := b.Build(nil); !errors.Is(e, goannoy.ErrInvalidOption) {
		t.Fatal(e)
	}
	if _, e := b.BuildTo(nil, path); !errors.Is(e, goannoy.ErrInvalidOption) {
		t.Fatal(e)
	}
}
func TestExportedFloatAPIs(t *testing.T) {
	t.Run("angular", func(t *testing.T) {
		exerciseFloat[*angular.Builder, *angular.Index, *angular.Searcher](t, angular.NewBuilder)
	})
	t.Run("dot", func(t *testing.T) { exerciseFloat[*dot.Builder, *dot.Index, *dot.Searcher](t, dot.NewBuilder) })
	t.Run("euclidean", func(t *testing.T) {
		exerciseFloat[*euclidean.Builder, *euclidean.Index, *euclidean.Searcher](t, euclidean.NewBuilder)
	})
	t.Run("manhattan", func(t *testing.T) {
		exerciseFloat[*manhattan.Builder, *manhattan.Index, *manhattan.Searcher](t, manhattan.NewBuilder)
	})
}
