package interfaces

type Random interface {
	Next() uint64
	NextSide() Side
	// NextIndex draw random integer between 0 and n-1 where n is at most the number of data points you have
	NextIndex(n ItemID) ItemID
	SetSeed(seed uint64)
	GetSeed() uint64
	// CloneAndReset returns a cloned instance and that is reset with the original seed.
	CloneAndReset() Random
}
