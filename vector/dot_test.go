package vector

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"unsafe"
)

func TestDotFloat32MatchesScalar(t *testing.T) {
	lengths := []int{0, 1, 3, 7, 8, 15, 16, 17, 20, 31, 32, 33, 64, 100, 255, 768, 1536, 4096}
	rng := rand.New(rand.NewSource(42))

	for _, vectorLength := range lengths {
		t.Run(fmt.Sprintf("len_%d", vectorLength), func(t *testing.T) {
			for trial := 0; trial < 8; trial++ {
				a := make([]float32, vectorLength)
				b := make([]float32, vectorLength)

				for i := 0; i < vectorLength; i++ {
					a[i] = rng.Float32()*2 - 1
					b[i] = rng.Float32()*2 - 1
				}

				want := dotFloat32Scalar(a, b, vectorLength)
				got := Dot(a, b, vectorLength)

				if diff := math.Abs(float64(got - want)); diff > dotFloatTolerance(vectorLength) {
					t.Fatalf("Dot mismatch for len=%d: got=%v want=%v diff=%v backend=%s",
						vectorLength, got, want, diff, activeDotFloat32BackendName())
				}

				gotUnsafe := DotUnsafe(
					unsafe.SliceData(a),
					unsafe.SliceData(b),
					vectorLength,
				)

				if diff := math.Abs(float64(gotUnsafe - want)); diff > dotFloatTolerance(vectorLength) {
					t.Fatalf("DotUnsafe mismatch for len=%d: got=%v want=%v diff=%v backend=%s",
						vectorLength, gotUnsafe, want, diff, activeDotFloat32BackendName())
				}
			}
		})
	}
}

func TestDotFloat64MatchesScalar(t *testing.T) {
	a := []float64{0.5, -1.25, 2.5, -3.75, 4.125}
	b := []float64{-2, 0.75, 1.5, -0.25, 3.0}

	want := dotScalar(a, b, len(a))
	got := Dot(a, b, len(a))

	if got != want {
		t.Fatalf("float64 Dot mismatch: got=%v want=%v", got, want)
	}

	gotUnsafe := DotUnsafe(
		unsafe.SliceData(a),
		unsafe.SliceData(b),
		len(a),
	)

	if gotUnsafe != want {
		t.Fatalf("float64 DotUnsafe mismatch: got=%v want=%v", gotUnsafe, want)
	}
}

func BenchmarkDotFloat32Backends(b *testing.B) {
	dims := []int{3, 10, 40, 128, 512, 1536}

	for _, dim := range dims {
		a, c := benchmarkFloat32Vectors(dim)
		aPtr := unsafe.SliceData(a)
		cPtr := unsafe.SliceData(c)

		b.Run(fmt.Sprintf("dim=%d/scalar", dim), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(dim * 2 * 4))
			for i := 0; i < b.N; i++ {
				_ = dotFloat32UnsafeScalar(aPtr, cPtr, dim)
			}
		})

		b.Run(fmt.Sprintf("dim=%d/active=%s", dim, activeDotFloat32BackendName()), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(dim * 2 * 4))
			for i := 0; i < b.N; i++ {
				_ = dotFloat32Unsafe(aPtr, cPtr, dim)
			}
		})
	}
}

func benchmarkFloat32Vectors(dim int) ([]float32, []float32) {
	rng := rand.New(rand.NewSource(int64(1000 + dim)))
	a := make([]float32, dim)
	b := make([]float32, dim)

	for i := 0; i < dim; i++ {
		a[i] = rng.Float32()*2 - 1
		b[i] = rng.Float32()*2 - 1
	}

	return a, b
}

func dotFloatTolerance(vectorLength int) float64 {
	if vectorLength <= 32 {
		return 1e-5
	}

	return 1e-4
}
