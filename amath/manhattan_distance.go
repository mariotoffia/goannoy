package amath

func ManhattanDistance[T Calculable](a, b []T) T {
	var sum T
	for i := range a {
		sum += Abs(a[i] - b[i])
	}
	return sum
}
