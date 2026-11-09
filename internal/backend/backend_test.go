package backend

import (
	"math"
	"math/rand"
	"testing"
)

func TestDotBackendsMatchReference(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	for _, n := range []int{0, 1, 3, 4, 7, 8, 15, 16, 17, 31, 32, 33, 127, 128, 129, 1536} {
		// Start at offset one to exercise unaligned vector loads.
		a, b := make([]float32, n+1), make([]float32, n+1)
		var want float64
		for j := 1; j <= n; j++ {
			a[j] = float32(r.NormFloat64())
			b[j] = float32(r.NormFloat64())
			want += float64(a[j]) * float64(b[j])
		}
		got := Dot(a[1:], b[1:])
		if math.Abs(float64(got)-want) > 2e-5*math.Max(1, math.Abs(want)) {
			t.Fatalf("n=%d got=%g want=%g", n, got, want)
		}
		if scalar := ScalarDot(a[1:], b[1:]); math.Abs(float64(scalar)-want) > 2e-5*math.Max(1, math.Abs(want)) {
			t.Fatalf("scalar n=%d got=%g want=%g", n, scalar, want)
		}
	}
}
