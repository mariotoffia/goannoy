package metric

import (
	"github.com/mariotoffia/goannoy/internal/codec"
	"math"
	"testing"
)

func TestMetricDegenerateVectorsMatchReference(t *testing.T) {
	for _, tc := range []struct {
		k    codec.Kind
		x, y []float32
		want float32
	}{
		{codec.Angular, []float32{0, 0}, []float32{0, 0}, 2},
		{codec.Angular, []float32{1, 0}, []float32{0, 1}, 2},
		{codec.Angular, []float32{1, 0}, []float32{1, 0}, 0},
		{codec.Dot, []float32{1, 2}, []float32{3, 4}, -11},
		{codec.Euclidean, []float32{1, 2}, []float32{4, 6}, 25},
		{codec.Manhattan, []float32{1, 2}, []float32{4, 6}, 7},
	} {
		l, _ := codec.New(tc.k, 2)
		b := make([]byte, l.Stride*2)
		copy(l.Floats(b, 0), tc.x)
		copy(l.Floats(b, 1), tc.y)
		Init(l, l.Node(b, 0))
		Init(l, l.Node(b, 1))
		if tc.k == codec.Dot {
			b[20] = 1
		}
		if got := Raw(l, l.Node(b, 0), l.Node(b, 1)); got != tc.want {
			t.Errorf("%s=%v want %v", tc.k, got, tc.want)
		}
	}
	if got := Normalize(codec.Angular, -2.3841858e-7); got != 0 || math.IsNaN(float64(got)) {
		t.Fatalf("negative angular normalization: %g", got)
	}
	l, _ := codec.New(codec.Angular, 2)
	for _, v := range [][]float32{{1}, {1, 2, 3}, {1, float32(math.NaN())}, {float32(math.Inf(1)), 0}} {
		if l.CheckFloats(v) == nil {
			t.Errorf("accepted %v", v)
		}
	}
}

// A dot plane must include the extra dimension in both normalization and side.
func TestDotFactorParticipatesInMargin(t *testing.T) {
	l, _ := codec.New(codec.Dot, 2)
	a := make([]byte, l.Stride)
	b := make([]byte, l.Stride)
	copy(l.Floats(a, 0), []float32{1, 0})
	codec.PutF32(a, 12, 2)
	copy(l.Floats(b, 0), []float32{-1, 0})
	codec.PutF32(b, 12, 1)
	if got := BuildMargin(l, a, b); got != 1 {
		t.Fatalf("margin=%g", got)
	}
}
