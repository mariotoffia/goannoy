package compat

import (
	"errors"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/angular"
	"github.com/mariotoffia/goannoy/dot"
	"github.com/mariotoffia/goannoy/euclidean"
	"github.com/mariotoffia/goannoy/hamming"
	"github.com/mariotoffia/goannoy/manhattan"
	"math"
	"testing"
)

func TestSearchSpotifyFixtures(t *testing.T) {
	query := []float32{1.25, -2.5, .75}
	m := readManifest(t)
	for _, f := range m.Fixtures {
		t.Run(f.Metric, func(t *testing.T) {
			dst := make([]goannoy.Result, 10)
			var n int
			var err error
			switch f.Metric {
			case "angular":
				i, e := angular.Open("fixtures/angular.ann", 3)
				if e != nil {
					t.Fatal(e)
				}
				defer i.Close()
				n, err = i.NewSearcher().SearchInto(query, dst, 10000)
			case "dot":
				i, e := dot.Open("fixtures/dot.ann", 3)
				if e != nil {
					t.Fatal(e)
				}
				defer i.Close()
				n, err = i.NewSearcher().SearchInto(query, dst, 10000)
			case "euclidean":
				i, e := euclidean.Open("fixtures/euclidean.ann", 3)
				if e != nil {
					t.Fatal(e)
				}
				defer i.Close()
				n, err = i.NewSearcher().SearchInto(query, dst, 10000)
			case "manhattan":
				i, e := manhattan.Open("fixtures/manhattan.ann", 3)
				if e != nil {
					t.Fatal(e)
				}
				defer i.Close()
				n, err = i.NewSearcher().SearchInto(query, dst, 10000)
			case "hamming":
				i, e := hamming.Open("fixtures/hamming.ann", 130)
				if e != nil {
					t.Fatal(e)
				}
				defer i.Close()
				n, err = i.NewSearcher().SearchInto([]uint64{0x123456789abcdef0, 0xfedcba9876543210, 1}, dst, 10000)
			}
			if err != nil {
				t.Fatal(err)
			}
			if n != 10 {
				t.Fatalf("n=%d", n)
			}
			for j, r := range dst {
				if int32(r.ID) != f.IDs[j] || math.Abs(float64(r.Distance-f.Distances[j])) > 1e-5 {
					t.Fatalf("[%d]=%+v want %d/%g", j, r, f.IDs[j], f.Distances[j])
				}
			}
		})
	}
}

func TestSearchRejectsInvalidPublicInput(t *testing.T) {
	i, e := angular.Open("fixtures/angular.ann", 3)
	if e != nil {
		t.Fatal(e)
	}
	defer i.Close()
	s := i.NewSearcher()
	dst := make([]goannoy.Result, 10)
	for _, v := range [][]float32{nil, {1, 2}, {1, 2, 3, 4}, {1, float32(math.NaN()), 2}} {
		if _, e := s.SearchInto(v, dst, -1); !errors.Is(e, goannoy.ErrInvalidVector) {
			t.Fatal(e)
		}
	}
	if _, e := s.SearchInto([]float32{1, 2, 3}, dst, -2); !errors.Is(e, goannoy.ErrInvalidSearchK) {
		t.Fatal(e)
	}
	if n, e := s.SearchInto([]float32{1, 2, 3}, nil, -1); e != nil || n != 0 {
		t.Fatalf("%d %v", n, e)
	}
	for _, id := range []goannoy.ItemID{-1, 64, math.MaxInt32} {
		if _, e := s.SearchByIDInto(id, dst, -1); !errors.Is(e, goannoy.ErrInvalidID) {
			t.Fatal(e)
		}
		if _, e := i.Vector(id); !errors.Is(e, goannoy.ErrInvalidID) {
			t.Fatal(e)
		}
		if _, e := i.Distance(0, id); !errors.Is(e, goannoy.ErrInvalidID) {
			t.Fatal(e)
		}
	}
	if _, e := s.Search([]float32{1, 2, 3}, -1, -1); !errors.Is(e, goannoy.ErrInvalidOption) {
		t.Fatal(e)
	}
	if e := i.Close(); e != nil {
		t.Fatal(e)
	}
	if _, e := s.SearchInto([]float32{1, 2, 3}, dst, -1); !errors.Is(e, goannoy.ErrClosed) {
		t.Fatal(e)
	}
}

func TestIndexVectorAndDistance(t *testing.T) {
	i, e := angular.Open("fixtures/angular.ann", 3)
	if e != nil {
		t.Fatal(e)
	}
	defer i.Close()
	v, e := i.Vector(0)
	if e != nil {
		t.Fatal(e)
	}
	if v[0] != -6 || v[1] != -5.5 || v[2] != -5.125 {
		t.Fatal(v)
	}
	v[0] = 123
	if e := i.VectorInto(0, v); e != nil || v[0] != -6 {
		t.Fatalf("%v %v", v, e)
	}
	d, e := i.Distance(0, 0)
	if e != nil || d > 1e-3 {
		t.Fatalf("%g %v", d, e)
	}
	r, e := i.NewSearcher().Search(v, 1, 10000)
	if e != nil || len(r) != 1 || r[0].ID != 0 {
		t.Fatalf("%v %v", r, e)
	}
	r = make([]goannoy.Result, 1)
	if n, e := i.NewSearcher().SearchByIDInto(0, r, 10000); e != nil || n != 1 || r[0].ID != 0 {
		t.Fatalf("%v %v", r, e)
	}
}
