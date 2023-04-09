package interfaces

type AnnoyIndex[
	TIdx int,
	TV VectorType,
	TRandType RandomTypes,
	TRand Random[TRandType]] interface {
}

type AnnoyIndexBuilder interface {
	ThreadBuild(treesPerWorker, workerIdx int, threadedBuildPolicy AnnoyIndexBuildPolicy)
}
