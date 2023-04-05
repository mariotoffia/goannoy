package vector

import "math"

// GetNorm normalizes the vector v?
func GetNorm[T Calculable](v []T, vectorLength int) T {
	return T(math.Sqrt(float64(Dot(v, v, vectorLength))))
}

func GetNormUnsafe[T Calculable](v *T, vectorLength int) T {
	return T(math.Sqrt(float64(DotUnsafe(v, v, vectorLength))))
}
