package interfaces

type AnnoyIndexBuildPolicy interface {
	Build(builder AnnoyIndexBuilder, numberOfTrees, nThreads int)
	LockNNodes()
	UnlockNNodes()
	LockNodes()
	UnlockNodes()
	LockSharedNodes()
	UnlockSharedNodes()
	LockRoots()
	UnlockRoots()
}
