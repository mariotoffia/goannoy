//go:build avx256
// +build avx256

package platform

import "fmt"

// #include <immintrin.h>
import "C"

func init() {
	DotProduct = dotF32AVX256
}

//go:noescape
func avxDotProduct(result *float32, x, y *float32, v C.__m256)

func dotF32AVX256(x, y []float32) float32 {
	if len(x) != len(y) {
		panic("Vectors must have the same length")
	}

	var result float32
	avxDotProduct(&result, &x[0], &y[0], len(x))
	return result
}

func test_avx256() {
	x := []float32{1, 2, 3, 4, 5, 6, 7, 8}
	y := []float32{1, 1, 1, 1, 1, 1, 1, 1}
	fmt.Printf("Dot product: %f\n", dotF32AVX256(x, y))
}
