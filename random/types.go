package random

type RandomTypes interface {
	uint32 | uint64
}

type Random[T RandomTypes] interface {
	Next() T
	NextBool() bool
	// NextIndex draw random integer between 0 and n-1 where n is at most the number of data points you have
	NextIndex(n T) T
	SetSeed(seed T)
}
