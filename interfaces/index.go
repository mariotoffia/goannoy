package interfaces

type AnnoyIndex[
	TIdx int,
	TV VectorType,
	TRandType RandomTypes,
	TRand Random[TRandType]] interface {
}

type AnnoyIndexBuilder interface {
	ThreadBuild(treesPerThread, threadIdx int, threadedBuildPolicy AnnoyIndexBuildPolicy)
}
