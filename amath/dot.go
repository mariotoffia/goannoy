package amath

func Dot[T Calculable](a, b []T) T {
	var sum T
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}
