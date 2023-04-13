package vector

import "github.com/mariotoffia/goannoy/interfaces"

func EuclideanDistance[TV interfaces.VectorType, TIX interfaces.IndexTypes](
	a, b []TV, vectorLength TIX,
) TV {
	var sum TV
	for i := TIX(0); i < vectorLength; i++ {
		sum += (a[i] - b[i]) * (a[i] - b[i])
	}
	return sum
}
