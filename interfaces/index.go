package interfaces

import (
	"io"
)

type IndexTypes interface {
	uint32 | uint64
}

type AnnoyIndexContext[TV VectorType, TIX IndexTypes] interface {
}

type AnnoyIndex[TV VectorType, TIX IndexTypes] interface {
	io.Closer
	// VectorLength returns the vector length of the index.
	VectorLength() TIX
	// GetItem returns the vector of the given _itemIndex_.
	GetItem(itemIndex TIX) []TV
	// AddItem adds an item to the index. The ownership of the vector _v_ is taken
	// by this function. The _itemIndex_ is a numbering index of the _v_ vector and
	// *SHOULD* be incremental. If same _itemIndex_ is added twice, the last one
	// will be the one in the index.
	AddItem(itemIndex TIX, v []TV)
	// Build will build a a new index. The _numberOfTrees_ is the number of trees
	// to build. The _numWorkers_ is the number of workers to use when building
	// the index. If _numWorkers_ is -1, the number of workers will be set to the
	// number of CPU cores. If _numWorkers_ is 0, the number of workers will be
	// set to 1. Hence, run on current goroutine.
	//
	// The _numberOfTrees_ will be split amongst the workers. The more number
	// of trees, the larger the index. But it also will be more precise.
	Build(numberOfTrees, numWorkers int)
	// CreateContext will create a batch context, that should be used in subsequent
	// calls to `GetNnsByVector` and `GetNnsByItem`.
	//
	// Create a new context per goroutine. Whenever a new index is loaded or saved,
	// a new context *must* be created since it contains vital information about the
	// index. Same applies when the index is *built!*
	CreateContext() AnnoyIndexContext[TV, TIX]
	// GetDistance returns the distance between the two given items.
	GetDistance(i, j TIX) TV
	// GetNnsByItem will search for the closest vectors to the given _item_ in the index. When
	// _numReturn_ is -1, it will search number of trees in index * _numReturn_.
	GetNnsByItem(
		item TIX,
		numReturn, numNodesToInspect int,
		ctx AnnoyIndexContext[TV, TIX],
	) (result []TIX, distances []TV)
	// GetAllNns will search for the closest vectors to the given _vector_. When
	// _numReturn_ is -1, it will search number of trees in index * _numReturn_.
	GetNnsByVector(
		vector []TV,
		numReturn, numNodesToInspect int,
		ctx AnnoyIndexContext[TV, TIX],
	) (result []TIX, distances []TV)
	Save(fileName string) error
	Load(fileName string) error
}

type AnnoyIndexBuilder interface {
	ThreadBuild(treesPerWorker, workerIdx int, threadedBuildPolicy AnnoyIndexBuildPolicy)
}
