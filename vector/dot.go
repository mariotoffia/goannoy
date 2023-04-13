package vector

import (
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
)

func Dot[TV interfaces.VectorType, TIX interfaces.IndexTypes](a, b []TV, vectorLength TIX) TV {
	var sum TV
	for i := TIX(0); i < vectorLength; i++ {
		sum += a[i] * b[i]
	}
	return sum
}

func DotUnsafe[TV interfaces.VectorType, TIX interfaces.IndexTypes](a, b *TV, vectorLength TIX) TV {
	a_ptr := unsafe.Pointer(a)
	b_ptr := unsafe.Pointer(b)
	size := TIX(unsafe.Sizeof(TV(0)))

	var sum TV

	for i := TIX(0); i < vectorLength; i++ {
		a := *(*TV)(unsafe.Pointer(unsafe.Add(a_ptr, i*size)))
		b := *(*TV)(unsafe.Pointer(unsafe.Add(b_ptr, i*size)))
		sum += a * b
	}

	return sum
}
