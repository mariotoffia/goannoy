//go:build accelerate && arm64

package vector

func init() {
	dotFloat32BackendName = "arm64-neon"
	dotFloat32BackendEnabled = true
	dotFloat32BackendUnsafeFn = dotFloat32AcceleratedUnsafeARM64
}

func dotFloat32AcceleratedUnsafeARM64(a, b *float32, vectorLength int) float32
