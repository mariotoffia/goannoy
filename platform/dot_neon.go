//go:build neon
// +build neon

package platform

func init() {
	DotProduct = dotF32Neon
}

//go:noescape
func neonDotProductF32(result *float32, x, y *float32, f uint32)

func dotF32Neon(x, y *float32, vectorLength uint32) float32 {

	var result float32
	neonDotProductF32(&result, x, y, vectorLength)
	return result
}
