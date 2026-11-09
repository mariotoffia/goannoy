package metric

import (
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/core"
	"math"
	"os"
	"testing"
)

// Expected planes were produced by the pinned C++ create_split at seed 42.
func TestSplitMatchesSpotify(t *testing.T) {
	for _, tc := range []struct {
		k     codec.Kind
		v     [3]float32
		extra float32
	}{
		{codec.Angular, [3]float32{.992901444, .118098907, -.0141234761}, 0},
		{codec.Dot, [3]float32{.959846556, .252543181, .0749453902}, .0964345112},
		{codec.Euclidean, [3]float32{.964479268, .258008987, .0566671006}, .284056544},
		{codec.Manhattan, [3]float32{.963474333, .260705858, .0612340756}, .254560411},
	} {
		l, _ := codec.New(tc.k, 3)
		b, err := os.ReadFile("../../compat/fixtures/" + tc.k.String() + ".ann")
		if err != nil {
			t.Fatal(err)
		}
		ids := make([]int32, 64)
		for i := range ids {
			ids[i] = int32(i)
			if tc.k == codec.Dot {
				l.Node(b, i)[20] = 0
			}
		}
		p := make([]byte, l.Stride)
		Split(l, b, ids, core.NewKiss64(42), p)
		for i, v := range l.Floats(p, 0) {
			if math.Abs(float64(v-tc.v[i])) > 1e-6 {
				t.Errorf("%s[%d]=%.9g want %.9g", tc.k, i, v, tc.v[i])
			}
		}
		off := 4
		if tc.k == codec.Dot {
			off = 12
		}
		if tc.k != codec.Angular && math.Abs(float64(codec.F32(p, off)-tc.extra)) > 1e-6 {
			t.Errorf("%s extra=%g", tc.k, codec.F32(p, off))
		}
	}
}

func TestHammingSplitUsesPackedBit(t *testing.T) {
	l, _ := codec.New(codec.Hamming, 130)
	b, err := os.ReadFile("../../compat/fixtures/hamming.ann")
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]int32, 64)
	for i := range ids {
		ids[i] = int32(i)
	}
	p := make([]byte, l.Stride)
	Split(l, b, ids, core.NewKiss64(42), p)
	count := 0
	for _, id := range ids {
		count += int(BuildMargin(l, p, l.Node(b, int(id))))
	}
	if count < 1 || count == 64 {
		t.Fatal("split does not partition packed vectors")
	}
}
