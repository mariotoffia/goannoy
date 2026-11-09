package core

import "testing"

func TestKissSequencesMatchSpotify(t *testing.T) {
	a := NewKiss32(42)
	b := NewKiss64(42)
	w32 := []uint32{555699124, 1451920892, 3647214204, 1646272917, 3123578926}
	w64 := []uint64{5680804532875375661, 421217147685971266, 15020224898898924438, 16951546275468841018, 12649982757840217320}
	for n := range w32 {
		if got := a.Next(); got != w32[n] {
			t.Fatalf("Kiss32[%d]=%d", n, got)
		}
		if got := b.Next(); got != w64[n] {
			t.Fatalf("Kiss64[%d]=%d", n, got)
		}
	}
}
