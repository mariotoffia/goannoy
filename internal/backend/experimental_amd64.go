//go:build amd64 && goannoy_experimental_simd && goexperiment.simd && !goannoy_scalar

package backend

import "simd/archsimd"

// The feature check happens once. Unsupported CPUs retain the scalar function.
var Dot = ScalarDot
var Name = "scalar"

func init() {
	if archsimd.X86.AVX() {
		Dot = dotSIMD
		Name = "amd64-experimental-avx"
	}
}
func dotSIMD(a, b []float32) float32 {
	var sum archsimd.Float32x8
	n := 0
	for ; n+8 <= len(a); n += 8 {
		sum = sum.Add(archsimd.LoadFloat32x8(a[n:]).Mul(archsimd.LoadFloat32x8(b[n:])))
	}
	var lanes [8]float32
	sum.Store(lanes[:])
	v := lanes[0] + lanes[1] + lanes[2] + lanes[3] + lanes[4] + lanes[5] + lanes[6] + lanes[7]
	for ; n < len(a); n++ {
		v += a[n] * b[n]
	}
	return v
}
