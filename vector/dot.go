package vector

import (
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
)

func Dot[TV interfaces.VectorType](a, b []TV, vectorLength int) TV {
	var sum TV
	for i := 0; i < vectorLength; i++ {
		sum += a[i] * b[i]
	}
	return sum
}

func DotUnsafe[TV interfaces.VectorType](a, b *TV, vectorLength int) TV {
	a_ptr := unsafe.Pointer(a)
	b_ptr := unsafe.Pointer(b)
	size := unsafe.Sizeof(TV(0))

	var sum TV

	for i := 0; i < vectorLength; i++ {
		offset := uintptr(i) * size
		a := *(*TV)(unsafe.Pointer(unsafe.Add(a_ptr, offset)))
		b := *(*TV)(unsafe.Pointer(unsafe.Add(b_ptr, offset)))
		sum += a * b
	}

	return sum
}
