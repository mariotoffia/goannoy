package interfaces

type AnnoyIndexBuildPolicy interface {
	Build(builder AnnoyIndexBuilder, treesPerThread int, nThreads int)
	LockNNodes()
	UnlockNNodes()
	LockNodes()
	UnlockNodes()
	LockSharedNodes()
	UnlockSharedNodes()
	LockRoots()
	UnlockRoots()
}
