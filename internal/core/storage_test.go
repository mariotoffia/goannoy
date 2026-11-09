package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteFailurePreservesDestination(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.ann")
	if e := os.WriteFile(path, []byte("old"), 0600); e != nil {
		t.Fatal(e)
	}
	failure := errors.New("write interrupted")
	if e := AtomicWrite(path, func(f *os.File) error {
		if _, e := f.WriteString("partial"); e != nil {
			return e
		}
		return failure
	}); !errors.Is(e, failure) {
		t.Fatal(e)
	}
	b, e := os.ReadFile(path)
	if e != nil || string(b) != "old" {
		t.Fatalf("%s %v", b, e)
	}
	if e := AtomicWrite(path, func(f *os.File) error { return f.Close() }); e == nil {
		t.Fatal("sync of closed file did not fail")
	}
	files, e := filepath.Glob(filepath.Join(dir, ".goannoy-*"))
	if e != nil || len(files) != 0 {
		t.Fatalf("%v %v", files, e)
	}
	if e := AtomicWrite(path, func(f *os.File) error { _, e := f.WriteString("new"); return e }); e != nil {
		t.Fatal(e)
	}
	b, e = os.ReadFile(path)
	if e != nil || string(b) != "new" {
		t.Fatalf("%s %v", b, e)
	}
}
