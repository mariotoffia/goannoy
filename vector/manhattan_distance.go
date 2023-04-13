package vector

import "github.com/mariotoffia/goannoy/interfaces"

func ManhattanDistance[T interfaces.VectorType, TIX interfaces.IndexTypes](
	a, b []T, vectorLength TIX,
) T {
	var sum T
	for i := TIX(0); i < vectorLength; i++ {
		sum += Abs(a[i] - b[i])
	}
	return sum
}
