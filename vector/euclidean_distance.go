package vector

func EuclideanDistance[T Calculable](a, b []T, vectorLength int) T {
	var sum T
	for i := 0; i < vectorLength; i++ {
		sum += (a[i] - b[i]) * (a[i] - b[i])
	}
	return sum
}
