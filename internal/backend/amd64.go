//go:build amd64 && !goannoy_scalar && (!goannoy_experimental_simd || !goexperiment.simd)

package backend

const Name = "amd64-sse2"

//go:noescape
func dotAsm(a, b *float32, n int) float32
func Dot(a, b []float32) float32 {
	if len(a) == 0 {
		return 0
	}
	return dotAsm(&a[0], &b[0], len(a))
}
