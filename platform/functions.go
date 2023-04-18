package platform

type DotProductFunc func(x, y *float32, vectorLength uint32) float32

var DotProduct DotProductFunc
