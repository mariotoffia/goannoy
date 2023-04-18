//go:build !avx256 && !avx512 && !neon
// +build !avx256,!avx512,!neon

package platform

func init() {
	DotProduct = dotF32Native
}

func dotF32Native(x, y []float32, vectorLength int) float32 {
	var sum float32
	for i := 0; i < vectorLength; i++ {
		sum += x[i] * y[i]
	}
	return sum
}
