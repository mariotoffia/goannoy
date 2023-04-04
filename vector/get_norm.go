package vector

import "math"

// GetNorm normalizes the vector v?
func GetNorm[T Calculable](v [ANNOYLIB_V_ARRAY_SIZE]T, vectorLength int) T {
	return T(math.Sqrt(float64(Dot(v, v, vectorLength))))
}
