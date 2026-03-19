package vector

import "github.com/mariotoffia/goannoy/interfaces"

func EuclideanDistance[TV interfaces.VectorType](
	a, b []TV, vectorLength int,
) TV {
	var sum TV
	for i := 0; i < vectorLength; i++ {
		sum += (a[i] - b[i]) * (a[i] - b[i])
	}
	return sum
}
