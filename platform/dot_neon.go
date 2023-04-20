//go:build neon
// +build neon

package platform

func init() {
	DotProduct = dotF32Neon
}

//go:noescape
func neonDotProductF32(v1, v2 *float32, length int) float32

func dotF32Neon(x, y *float32, vectorLength uint32) float32 {
	return neonDotProductF32(x, y, int(vectorLength))
}
