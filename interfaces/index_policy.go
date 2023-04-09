package interfaces

type AnnoyIndexBuildPolicy interface {
	// Build will build a a new index. The _numberOfTrees_ is the number of trees
	// to build. The _numWorkers_ is the number of workers to use when building
	// the index. If _numWorkers_ is -1, the number of workers will be set to the
	// number of CPU cores. If _numWorkers_ is 0, the number of workers will be
	// set to 1. Hence, run on current goroutine.
	//
	// The _numberOfTrees_ will be split amongst the workers. The more number
	// of trees, the larger the index. But it also will be more precise.
	//
	// This uses the `AnnoyIndexBuilder` to perform the actual work.
	Build(builder AnnoyIndexBuilder, numberOfTrees, numberOfWorker int)
	LockNNodes()
	UnlockNNodes()
	LockNodes()
	UnlockNodes()
	LockSharedNodes()
	UnlockSharedNodes()
	LockRoots()
	UnlockRoots()
}
