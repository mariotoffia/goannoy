//go:build neon
// +build neon

package platform

import _ "unsafe" // Required for go:linkname

func init() {
	DotProduct = dotF32Neon
}

//go:linkname neonDotProductF32 neonDotProductF32
func neonDotProductF32(v1, v2 *float32, length uint32) float32

func dotF32Neon(x, y *float32, vectorLength uint32) float32 {
	return neonDotProductF32(x, y, vectorLength)
}
