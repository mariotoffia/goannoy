package interfaces

type IndexTypes interface {
	uint32 | uint64
}

type AnnoyIndex[
	TIX IndexTypes,
	TV VectorType,
	TRand Random[TIX]] interface {
}

type AnnoyIndexBuilder interface {
	ThreadBuild(treesPerWorker, workerIdx int, threadedBuildPolicy AnnoyIndexBuildPolicy)
}
