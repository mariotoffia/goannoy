package engine

import (
	"errors"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildToPreservesExistingMappedSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.ann")
	b, _ := NewBuilder(codec.Angular, 2)
	b.Config.Workers(1).Trees(2)
	if e := b.AddFloat(0, []float32{1, 0}); e != nil {
		t.Fatal(e)
	}
	old, e := b.BuildTo(nil, path)
	if e != nil {
		t.Fatal(e)
	}
	defer old.Close()
	if e := b.AddFloat(0, []float32{0, 1}); e != nil {
		t.Fatal(e)
	}
	next, e := b.BuildTo(nil, path)
	if e != nil {
		t.Fatal(e)
	}
	defer next.Close()
	v := make([]float32, 2)
	if e := old.VectorFloat(0, v); e != nil || v[0] != 1 {
		t.Fatalf("existing mapping changed: %v %v", v, e)
	}
	if e := next.VectorFloat(0, v); e != nil || v[1] != 1 {
		t.Fatalf("new mapping wrong: %v %v", v, e)
	}
}

func TestOnDiskBuildAndCleanup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")
	b, _ := NewBuilder(codec.Angular, 3)
	b.Config.OnDisk(path).Trees(3)
	for id := 0; id < 64; id++ {
		if e := b.AddFloat(goannoy.ItemID(id), fixtureFloat(id)); e != nil {
			t.Fatal(e)
		}
	}
	i, e := b.Build(nil)
	if e != nil {
		t.Fatal(e)
	}
	defer i.Close()
	if _, e := os.Stat(path); e != nil {
		t.Fatal(e)
	}
	files, e := filepath.Glob(filepath.Join(dir, ".goannoy-*"))
	if e != nil || len(files) != 0 {
		t.Fatalf("%v %v", files, e)
	}
	if _, e := b.BuildTo(nil, filepath.Join(dir, "missing", "file.ann")); !errors.Is(e, os.ErrNotExist) {
		t.Fatal(e)
	}
}
