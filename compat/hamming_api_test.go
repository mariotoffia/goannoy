package compat

import (
	"errors"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/hamming"
	"path/filepath"
	"testing"
)

func TestExportedHammingAPI(t *testing.T) {
	if _, e := hamming.NewBuilder(0); !errors.Is(e, goannoy.ErrInvalidDimension) {
		t.Fatal(e)
	}
	b, e := hamming.NewBuilder(130)
	if e != nil {
		t.Fatal(e)
	}
	b.Trees(3).Workers(2).Seed(42).Capacity(32).Profile(goannoy.LowMemory)
	for id := 0; id < 32; id++ {
		if e := b.Add(goannoy.ItemID(id), []uint64{uint64(id), 0, 0}); e != nil {
			t.Fatal(e)
		}
	}
	i, e := b.Build(nil)
	if e != nil {
		t.Fatal(e)
	}
	if i.Dimension() != 130 || i.Items() != 32 || i.Trees() != 3 {
		t.Fatal("metadata")
	}
	v, e := i.Vector(7)
	if e != nil || v[0] != 7 {
		t.Fatalf("%v %v", v, e)
	}
	if e := i.VectorInto(7, v); e != nil {
		t.Fatal(e)
	}
	if e := i.VectorInto(7, nil); !errors.Is(e, goannoy.ErrInvalidVector) {
		t.Fatal(e)
	}
	if _, e := i.Vector(-1); !errors.Is(e, goannoy.ErrInvalidID) {
		t.Fatal(e)
	}
	if d, e := i.Distance(0, 7); e != nil || d != 3 {
		t.Fatalf("%g %v", d, e)
	}
	s := i.NewSearcher()
	dst := make([]goannoy.Result, 10)
	if n, e := s.SearchByIDInto(7, dst, 10000); e != nil || n != 10 || dst[0].ID != 7 {
		t.Fatalf("%d %v", n, e)
	}
	if r, e := s.Search(v, 10, 10000); e != nil || len(r) != 10 {
		t.Fatalf("%v %v", r, e)
	}
	if _, e := s.Search(v, -1, -1); !errors.Is(e, goannoy.ErrInvalidOption) {
		t.Fatal(e)
	}
	if _, e := s.Search(nil, 10, -1); !errors.Is(e, goannoy.ErrInvalidVector) {
		t.Fatal(e)
	}
	if e := i.Close(); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Search(v, 10, -1); !errors.Is(e, goannoy.ErrClosed) {
		t.Fatal(e)
	}
	if e := i.VectorInto(0, v); !errors.Is(e, goannoy.ErrClosed) {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "index.ann")
	b.OnDisk(path)
	j, e := b.Build(nil)
	if e != nil {
		t.Fatal(e)
	}
	if e := j.Close(); e != nil {
		t.Fatal(e)
	}
	j, e = b.BuildTo(nil, path)
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
	if _, e := b.Build(nil); !errors.Is(e, goannoy.ErrInvalidOption) {
		t.Fatal(e)
	}
	if _, e := b.BuildTo(nil, path); !errors.Is(e, goannoy.ErrInvalidOption) {
		t.Fatal(e)
	}
}
