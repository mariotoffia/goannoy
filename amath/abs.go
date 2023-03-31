package amath

import "math"

// Abs32 is the same as math.Abs, but for float32.
func Abs32(x float32) float32 {
	return math.Float32frombits(math.Float32bits(x) &^ (1 << 31))
}

// Abs is less performant on hardware where floating point operations are
// expensive, but it is generic.
func Abs[T Calculable](x T) T {
	if x < 0 {
		return -x
	}

	return x
}
