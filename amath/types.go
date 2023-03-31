package amath

type Calculable interface {
	// The allowed types that may be used in a calculation
	float32 | float64 | int32 | int64
}
