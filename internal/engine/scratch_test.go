package engine

import (
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"testing"
)

func TestScratchCapacityStartsIndependentOfIndex(t *testing.T) {
	i, e := Open("../../compat/fixtures/angular.ann", codec.Angular, 3)
	if e != nil {
		t.Fatal(e)
	}
	defer i.Close()
	s := i.NewSearcher()
	if cap(s.ids) != 0 || cap(s.queue) != 0 {
		t.Fatal("scratch preallocated from index")
	}
	dst := make([]goannoy.Result, 1)
	if _, e = s.SearchFloat([]float32{1, 2, 3}, dst, 1); e != nil {
		t.Fatal(e)
	}
	if cap(s.ids) > 64 || cap(s.queue) > 64 {
		t.Fatalf("scratch too large: %d/%d", cap(s.ids), cap(s.queue))
	}
}
