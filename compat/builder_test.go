package compat

import (
	"errors"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/angular"
	"github.com/mariotoffia/goannoy/hamming"
	"testing"
)

func TestFluentBuilder(t *testing.T) {
	b, e := angular.NewBuilder(2)
	if e != nil {
		t.Fatal(e)
	}
	b.Trees(3).Workers(1).Seed(42).Capacity(20).Profile(goannoy.LowMemory)
	if e := b.Err(); e != nil {
		t.Fatal(e)
	}
	if e := b.Add(0, []float32{1, 0}); e != nil {
		t.Fatal(e)
	}
	i, e := b.Build(nil)
	if e != nil {
		t.Fatal(e)
	}
	defer i.Close()
	if i.Items() != 1 || i.Trees() != 3 {
		t.Fatalf("%d %d", i.Items(), i.Trees())
	}
	b.Trees(0)
	if e := b.Add(1, []float32{0, 1}); !errors.Is(e, goannoy.ErrInvalidOption) {
		t.Fatal(e)
	}
	if _, e := b.Build(nil); !errors.Is(e, goannoy.ErrInvalidOption) {
		t.Fatal(e)
	}
	if _, e := angular.NewBuilder(0); !errors.Is(e, goannoy.ErrInvalidDimension) {
		t.Fatal(e)
	}
	h, e := hamming.NewBuilder(130)
	if e != nil {
		t.Fatal(e)
	}
	if e = h.Add(0, []uint64{1, 0, 4}); !errors.Is(e, goannoy.ErrInvalidVector) {
		t.Fatal(e)
	}
	if e = h.Add(0, []uint64{1, 0, 3}); e != nil {
		t.Fatal(e)
	}
}
