package compat

import (
	"errors"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/angular"
	"github.com/mariotoffia/goannoy/dot"
	"github.com/mariotoffia/goannoy/euclidean"
	"github.com/mariotoffia/goannoy/hamming"
	"github.com/mariotoffia/goannoy/manhattan"
	"os"
	"path/filepath"
	"testing"
)

func TestPublicOpenSpotifyFixtures(t *testing.T) {
	closers := []interface{ Close() error }{}
	a, e := angular.Open("fixtures/angular.ann", 3, goannoy.Prefault)
	if e != nil {
		t.Fatal(e)
	}
	closers = append(closers, a)
	d, e := dot.Open("fixtures/dot.ann", 3)
	if e != nil {
		t.Fatal(e)
	}
	closers = append(closers, d)
	eu, e := euclidean.Open("fixtures/euclidean.ann", 3)
	if e != nil {
		t.Fatal(e)
	}
	closers = append(closers, eu)
	m, e := manhattan.Open("fixtures/manhattan.ann", 3)
	if e != nil {
		t.Fatal(e)
	}
	closers = append(closers, m)
	h, e := hamming.Open("fixtures/hamming.ann", 130)
	if e != nil {
		t.Fatal(e)
	}
	closers = append(closers, h)
	if a.Items() != 64 || a.Trees() != 3 || a.Dimension() != 3 {
		t.Fatal("metadata")
	}
	for _, c := range closers {
		if e := c.Close(); e != nil {
			t.Fatal(e)
		}
		if e := c.Close(); e != nil {
			t.Fatal(e)
		}
	}
}
func TestPublicOpenErrors(t *testing.T) {
	if _, e := angular.Open("missing", 0); !errors.Is(e, goannoy.ErrInvalidDimension) {
		t.Fatal(e)
	}
	if _, e := angular.Open("missing", 3, 255); !errors.Is(e, goannoy.ErrInvalidOption) {
		t.Fatal(e)
	}
	if _, e := angular.Open("missing", 3); !errors.Is(e, os.ErrNotExist) {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "empty.ann")
	if e := os.WriteFile(path, nil, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := angular.Open(path, 3); !errors.Is(e, goannoy.ErrCorruptIndex) {
		t.Fatal(e)
	}
	if e := os.WriteFile(path, []byte{1, 2, 3}, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := angular.Open(path, 3); !errors.Is(e, goannoy.ErrCorruptIndex) {
		t.Fatal(e)
	}
}
