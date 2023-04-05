package vector

import "unsafe"

func Dot[T Calculable](a, b []T, vectorLength int) T {
	var sum T
	for i := 0; i < vectorLength; i++ {
		sum += a[i] * b[i]
	}
	return sum
}

func DotUnsafe[T Calculable](a, b *T, vectorLength int) T {
	a_ptr := unsafe.Pointer(a)
	b_ptr := unsafe.Pointer(b)
	size := int(unsafe.Sizeof(T(0)))

	var sum T

	for i := 0; i < vectorLength; i++ {
		a := *(*T)(unsafe.Pointer(unsafe.Add(a_ptr, i*size)))
		b := *(*T)(unsafe.Pointer(unsafe.Add(b_ptr, i*size)))
		sum += a * b
	}

	return sum
}
