//go:build neon
// +build neon

package platform

import "testing"

func TestNeonDot(t *testing.T) {
	y := []float32{4, 5, 6}
	x := []float32{1, 2, 3}
	dotF32Neon(&x[0], &y[0], 3)
}
