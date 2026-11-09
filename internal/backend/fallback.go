//go:build (!arm64 && !amd64) || goannoy_scalar

package backend

const Name = "scalar"

func Dot(a, b []float32) float32 { return ScalarDot(a, b) }
