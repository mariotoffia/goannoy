// Package backend provides stable platform vector kernels.
package backend

func ScalarDot(a, b []float32) float32 {
	// Establish the common bound once so the loop has no second slice check.
	b = b[:len(a)]
	var v float32
	n := 0
	for ; n+3 < len(a); n += 4 {
		x, y := a[n:n+4], b[n:n+4]
		// Keep the sequential float32 accumulation order, unlike SIMD reductions.
		v += float32(x[0] * y[0])
		v += float32(x[1] * y[1])
		v += float32(x[2] * y[2])
		v += float32(x[3] * y[3])
	}
	for ; n < len(a); n++ {
		v += float32(a[n] * b[n])
	}
	return v
}
