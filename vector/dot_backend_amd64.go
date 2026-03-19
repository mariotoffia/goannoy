//go:build accelerate && amd64 && goexperiment.simd

package vector

import (
	"simd/archsimd"
	"unsafe"
)

func init() {
	if !archsimd.X86.AVX() {
		return
	}

	dotFloat32BackendName = "amd64-archsimd"
	dotFloat32BackendEnabled = true
	dotFloat32BackendUnsafeFn = dotFloat32AcceleratedUnsafeAMD64
}

func dotFloat32AcceleratedUnsafeAMD64(a, b *float32, vectorLength int) float32 {
	if vectorLength <= 0 {
		return 0
	}

	return dotFloat32AcceleratedAMD64(
		unsafe.Slice(a, vectorLength),
		unsafe.Slice(b, vectorLength),
		vectorLength,
	)
}

func dotFloat32AcceleratedAMD64(a, b []float32, vectorLength int) float32 {
	var sum archsimd.Float32x8
	i := 0

	for ; i <= vectorLength-8; i += 8 {
		va := archsimd.LoadFloat32x8Slice(a[i:])
		vb := archsimd.LoadFloat32x8Slice(b[i:])
		sum = sum.Add(va.Mul(vb))
	}

	var tmp [8]float32
	sum.Store(&tmp)

	result := tmp[0] + tmp[1] + tmp[2] + tmp[3] +
		tmp[4] + tmp[5] + tmp[6] + tmp[7]

	for ; i < vectorLength; i++ {
		result += a[i] * b[i]
	}

	return result
}
