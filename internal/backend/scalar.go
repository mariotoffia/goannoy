// Package backend provides stable platform vector kernels.
package backend

func ScalarDot(a, b []float32) float32 {
	var v float32
	for n, x := range a {
		v += float32(x * b[n])
	}
	return v
}
