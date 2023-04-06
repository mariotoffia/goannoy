package interfaces

type RandomTypes interface {
	uint32 | uint64
}

type Random[T RandomTypes] interface {
	Next() T
	NextSide() Side
	// NextIndex draw random integer between 0 and n-1 where n is at most the number of data points you have
	NextIndex(n T) T
	SetSeed(seed T)
	GetSeed() T
	// CloneAndReset returns a cloned instance and that is reset with the original seed.
	CloneAndReset() Random[T]
}
