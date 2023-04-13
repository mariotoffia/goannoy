package vector

import (
	"math"

	"github.com/mariotoffia/goannoy/interfaces"
)

// GetNorm normalizes the vector v?
func GetNorm[T interfaces.VectorType, TIX interfaces.IndexTypes](v []T, vectorLength TIX) T {
	return T(math.Sqrt(float64(Dot(v, v, vectorLength))))
}

func GetNormUnsafe[T interfaces.VectorType, TIX interfaces.IndexTypes](v *T, vectorLength TIX) T {
	return T(math.Sqrt(float64(DotUnsafe(v, v, vectorLength))))
}
