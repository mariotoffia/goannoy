package interfaces

import (
	"io"
)

type ItemID = int32

type AnnoyIndexContext[TV VectorType] interface {
}

type AnnoyIndex[TV VectorType] interface {
	io.Closer
	// VectorLength returns the vector length of the index.
	VectorLength() int
	// GetItem returns the vector of the given _itemIndex_.
	GetItem(itemIndex ItemID) []TV
	// AddItem adds an item to the index. The ownership of the vector _v_ is taken
	// by this function. The _itemIndex_ is a numbering index of the _v_ vector and
	// *SHOULD* be incremental. If same _itemIndex_ is added twice, the last one
	// will be the one in the index.
	AddItem(itemIndex ItemID, v []TV) error
	// Build will build a a new index. The _numberOfTrees_ is the number of trees
	// to build. The _numWorkers_ is the number of workers to use when building
	// the index. If _numWorkers_ is -1, the number of workers will be set to the
	// number of CPU cores. If _numWorkers_ is 0, the number of workers will be
	// set to 1. Hence, run on current goroutine.
	//
	// The _numberOfTrees_ will be split amongst the workers. The more number
	// of trees, the larger the index. But it also will be more precise.
	Build(numberOfTrees, numWorkers int) error
	// CreateContext will create a batch context, that should be used in subsequent
	// calls to `GetNnsByVector` and `GetNnsByItem`.
	//
	// Create a new context per goroutine. Whenever a new index is loaded or saved,
	// a new context *must* be created since it contains vital information about the
	// index. Same applies when the index is *built!*
	CreateContext() AnnoyIndexContext[TV]
	// GetDistance returns the distance between the two given items.
	GetDistance(i, j ItemID) TV
	// GetNnsByItem will search for the closest vectors to the given _item_ in the index.
	// When _numNodesToInspect_ is -1, it will search number of trees in index * _numReturn_.
	GetNnsByItem(
		item ItemID,
		numReturn, numNodesToInspect int,
		ctx AnnoyIndexContext[TV],
	) (result []ItemID, distances []TV)
	// GetNnsByVector will search for the closest vectors to the given _vector_.
	// When _numNodesToInspect_ is -1, it will search number of trees in index * _numReturn_.
	GetNnsByVector(
		vector []TV,
		numReturn, numNodesToInspect int,
		ctx AnnoyIndexContext[TV],
	) (result []ItemID, distances []TV)
	// GetNnsByItemIDs performs the same search as GetNnsByItem but skips distance normalization/output.
	GetNnsByItemIDs(
		item ItemID,
		numReturn, numNodesToInspect int,
		ctx AnnoyIndexContext[TV],
	) []ItemID
	// GetNnsByVectorIDs performs the same search as GetNnsByVector but skips distance normalization/output.
	GetNnsByVectorIDs(
		vector []TV,
		numReturn, numNodesToInspect int,
		ctx AnnoyIndexContext[TV],
	) []ItemID
	// OnDiskBuild prepares the index to build directly into the given file.
	// All subsequent AddItem and Build calls write to a memory-mapped file
	// instead of heap memory. After Build completes the file is a ready-to-use
	// index — no separate Save call is needed. Must be called before AddItem.
	OnDiskBuild(fileName string) error
	// Unbuild removes all trees from the index, allowing new items to be added
	// and the index to be rebuilt. Items already added are preserved. This is
	// not valid on a loaded index — only on one that was built in memory.
	Unbuild() error
	// Unload releases the index data from memory but keeps the configuration
	// intact so that Load can be called again. This is useful for long-lived
	// applications that need to free memory without destroying the index object.
	Unload() error
	// GetNItems returns the number of items that have been added to the index.
	GetNItems() int
	// GetNTrees returns the number of trees in the index. This is only valid
	// after Build or Load has been called.
	GetNTrees() int
	Save(fileName string) error
	Load(fileName string) error
	// Prefault pre-loads all memory-mapped pages into RAM so that subsequent
	// searches do not trigger page faults. This is a no-op when the index is
	// not backed by a memory-mapped file. Call after Load for best effect.
	Prefault() error
}

type AnnoyIndexBuilder interface {
	ThreadBuild(treesPerWorker, workerIdx int, threadedBuildPolicy AnnoyIndexBuildPolicy)
}
