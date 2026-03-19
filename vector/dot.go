package vector

import (
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
)

const dotFloat32MinAcceleratedLength = 16

var (
	dotFloat32BackendName     = "scalar"
	dotFloat32BackendEnabled  bool
	dotFloat32BackendUnsafeFn = dotFloat32UnsafeScalar
)

func Dot[TV interfaces.VectorType](a, b []TV, vectorLength int) TV {
	var zero TV

	switch any(zero).(type) {
	case float32:
		return TV(dotFloat32Unsafe(
			(*float32)(unsafe.Pointer(unsafe.SliceData(a))),
			(*float32)(unsafe.Pointer(unsafe.SliceData(b))),
			vectorLength,
		))
	default:
		return dotScalar(a, b, vectorLength)
	}
}

func DotUnsafe[TV interfaces.VectorType](a, b *TV, vectorLength int) TV {
	var zero TV

	switch any(zero).(type) {
	case float32:
		return TV(dotFloat32Unsafe(
			(*float32)(unsafe.Pointer(a)),
			(*float32)(unsafe.Pointer(b)),
			vectorLength,
		))
	default:
		return dotUnsafeScalar(a, b, vectorLength)
	}
}

func dotScalar[TV interfaces.VectorType](a, b []TV, vectorLength int) TV {
	var sum TV
	for i := 0; i < vectorLength; i++ {
		sum += a[i] * b[i]
	}
	return sum
}

func dotUnsafeScalar[TV interfaces.VectorType](a, b *TV, vectorLength int) TV {
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

func activeDotFloat32BackendName() string {
	return dotFloat32BackendName
}

func dotFloat32Scalar(a, b []float32, vectorLength int) float32 {
	var sum float32

	for i := 0; i < vectorLength; i++ {
		sum += a[i] * b[i]
	}

	return sum
}

func dotFloat32Unsafe(a, b *float32, vectorLength int) float32 {
	if vectorLength <= 0 {
		return 0
	}

	if !dotFloat32BackendEnabled || vectorLength < dotFloat32MinAcceleratedLength {
		return dotFloat32UnsafeScalar(a, b, vectorLength)
	}

	return dotFloat32BackendUnsafeFn(a, b, vectorLength)
}

func dotFloat32UnsafeScalar(a, b *float32, vectorLength int) float32 {
	a_ptr := unsafe.Pointer(a)
	b_ptr := unsafe.Pointer(b)
	const size = unsafe.Sizeof(float32(0))

	var sum float32

	for i := 0; i < vectorLength; i++ {
		offset := uintptr(i) * size
		av := *(*float32)(unsafe.Pointer(unsafe.Add(a_ptr, offset)))
		bv := *(*float32)(unsafe.Pointer(unsafe.Add(b_ptr, offset)))
		sum += av * bv
	}

	return sum
}
