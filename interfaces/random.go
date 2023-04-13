package interfaces

type Random[TIX IndexTypes] interface {
	Next() TIX
	NextSide() Side
	// NextIndex draw random integer between 0 and n-1 where n is at most the number of data points you have
	NextIndex(n TIX) TIX
	SetSeed(seed TIX)
	GetSeed() TIX
	// CloneAndReset returns a cloned instance and that is reset with the original seed.
	CloneAndReset() Random[TIX]
}
