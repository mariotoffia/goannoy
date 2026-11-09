package compat

import (
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/hamming"
	"testing"
)

func TestHammingApproximateMatchesSpotify(t *testing.T) {
	i, e := hamming.Open("fixtures/hamming.ann", 130)
	if e != nil {
		t.Fatal(e)
	}
	defer i.Close()
	r, e := i.NewSearcher().Search([]uint64{0x123456789abcdef0, 0xfedcba9876543210, 1}, 10, goannoy.AutoSearchK)
	if e != nil {
		t.Fatal(e)
	}
	want := []goannoy.ItemID{2, 10, 61, 52, 48, 41, 58, 22, 35, 5}
	if len(r) != len(want) {
		t.Fatalf("%v", r)
	}
	for j, id := range want {
		if r[j].ID != id {
			t.Fatalf("result[%d]=%+v want ID %d", j, r[j], id)
		}
	}
}
