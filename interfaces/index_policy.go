package interfaces

type AnnoyIndexBuildPolicy interface {
	Build(builder AnnoyIndexBuilder, treesPerThread, nThreads int)
	LockNNodes()
	UnlockNNodes()
	LockNodes()
	UnlockNodes()
	LockSharedNodes()
	UnlockSharedNodes()
	LockRoots()
	UnlockRoots()
}
