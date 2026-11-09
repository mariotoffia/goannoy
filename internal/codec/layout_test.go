package codec

import (
	"errors"
	goannoy "github.com/mariotoffia/goannoy"
	"math"
	"os"
	"testing"
)

// Catches field offset and packed Hamming regressions against independent bytes.
func TestLayoutAndDecode(t *testing.T) {
	for _, tc := range []struct {
		k                          Kind
		dim, offset, stride, child int
	}{
		{Angular, 3, 12, 24, 4}, {Dot, 3, 24, 36, 4}, {Euclidean, 3, 16, 28, 8},
		{Manhattan, 3, 16, 28, 8}, {Hamming, 130, 16, 40, 4},
	} {
		l, err := New(tc.k, tc.dim)
		if err != nil {
			t.Fatal(err)
		}
		if l.VectorOffset != tc.offset || l.Stride != tc.stride || l.ChildrenOffset != tc.child {
			t.Fatalf("%+v", l)
		}
		b, err := os.ReadFile("../../compat/fixtures/" + tc.k.String() + ".ann")
		if err != nil {
			t.Fatal(err)
		}
		if l.Descendants(b, 0) != 1 {
			t.Fatal("item node is not a leaf")
		}
		if tc.k == Hamming {
			got := l.WordsAt(b, 2)
			if got[0] != 0x3c6ef372fe94f82a || got[2] != 2 {
				t.Fatalf("hamming %x", got)
			}
		} else if l.Floats(b, 0)[0] != -6 {
			t.Fatalf("vector %v", l.Floats(b, 0))
		}
		if tc.k == Dot && (F32(b, 12) != 0 || F32(b, l.Stride+12) <= 0 || b[20] != 1) {
			t.Fatal("dot factor/built layout")
		}
	}
	for _, k := range []Kind{Angular, Dot, Euclidean, Manhattan, Hamming} {
		for _, dim := range []int{1, 64, 65, 130} {
			if _, err := New(k, dim); err != nil {
				t.Fatal(err)
			}
		}
		for _, dim := range []int{0, -1, math.MaxInt} {
			if _, err := New(k, dim); !errors.Is(err, goannoy.ErrInvalidDimension) {
				t.Errorf("%d: %v", dim, err)
			}
		}
	}
}

func TestSplitNodeHasExactlyTwoChildren(t *testing.T) {
	l, _ := New(Angular, 3)
	b := make([]byte, l.Stride)
	PutI32(b, 0, 1000)
	PutI32(b, 4, 17)
	PutI32(b, 8, 23)
	got := l.Children(b, 0)
	if got != [2]int32{17, 23} {
		t.Fatalf("%v", got)
	}
}

func TestHammingRejectsDirtyTailBits(t *testing.T) {
	l, _ := New(Hamming, 130)
	for _, v := range [][]uint64{{0, 0}, {0, 0, 4}, {0, 0, 1 << 63}, {0, 0, 0, 0}} {
		if !errors.Is(l.CheckWords(v), goannoy.ErrInvalidVector) {
			t.Errorf("accepted %x", v)
		}
	}
	if err := l.CheckWords([]uint64{^uint64(0), 0, 3}); err != nil {
		t.Fatal(err)
	}
}
