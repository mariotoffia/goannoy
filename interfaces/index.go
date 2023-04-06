package interfaces

type AnnoyIndex[
	TIdx int,
	TV VectorType,
	TRandType RandomTypes,
	TRand Random[TRandType]] interface {
}

type AnnoyIndexBuilder interface {
	ThreadBuild(q int, threadIdx int, threadedBuildPolicy AnnoyIndexBuildPolicy)
}
