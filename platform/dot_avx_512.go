//go:build avx256
// +build avx256

package platform

import (
	"fmt"
)

func init() {
	DotProduct = dotF32AVX512
}

//go:noescape
func avxDotProductF32AVX512(result *float32, x, y *float32, f int)

func dotF32AVX512(x, y []float32) float32 {
	if len(x) != len(y) {
		panic("Vectors must have the same length")
	}

	var result float32
	avxDotProductF32AVX512(&result, &x[0], &y[0], len(x))
	return result
}

func test_avx512() {
	x := []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	y := []float32{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	fmt.Printf("Dot product (AVX512, float32): %f\n", dotF32AVX512(x, y))
}
