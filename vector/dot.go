package vector

func Dot[T Calculable](a, b []T, vectorLength int) T {
	var sum T
	for i := 0; i < vectorLength; i++ {
		sum += a[i] * b[i]
	}
	return sum
}
