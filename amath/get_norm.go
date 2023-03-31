package amath

import "math"

// GetNorm normalizes the vector v?
func GetNorm[T Calculable](v []T) T {
	return T(math.Sqrt(float64(Dot(v, v))))
}
