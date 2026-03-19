package builder

import (
	"github.com/mariotoffia/goannoy/distance/angular"
	"github.com/mariotoffia/goannoy/distance/dotproduct"
	"github.com/mariotoffia/goannoy/distance/euclidean"
	"github.com/mariotoffia/goannoy/distance/hamming"
	"github.com/mariotoffia/goannoy/distance/manhattan"
	"github.com/mariotoffia/goannoy/index"
	"github.com/mariotoffia/goannoy/index/memory"
	"github.com/mariotoffia/goannoy/index/policy"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/random"
)

type AnnoyIndexBuilderImpl[TV interfaces.VectorType] struct {
	allocHint            int
	random               interfaces.Random
	distance             interfaces.Distance[TV]
	buildPolicy          interfaces.AnnoyIndexBuildPolicy
	allocator            interfaces.BuildIndexAllocator
	indexMemoryAllocator interfaces.IndexAllocator
	sorter               interfaces.Sorter[TV]
	logVerbose           bool
}

// Index creates the default public builder shape used throughout this repo.
func Index() *AnnoyIndexBuilderImpl[float32] {
	return &AnnoyIndexBuilderImpl[float32]{}
}

// IndexOf creates a typed `AnnoyIndexBuilderImpl` instance for internal/advanced use.
func IndexOf[TV interfaces.VectorType]() *AnnoyIndexBuilderImpl[TV] {
	return &AnnoyIndexBuilderImpl[TV]{}
}

func (bld *AnnoyIndexBuilderImpl[TV]) Random(rnd interfaces.Random) *AnnoyIndexBuilderImpl[TV] {
	bld.random = rnd
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV]) IndexNumHint(allocHint int) *AnnoyIndexBuilderImpl[TV] {
	if allocHint <= 0 {
		return bld
	}

	bld.allocHint = allocHint
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV]) AngularDistance(vectorLength int) *AnnoyIndexBuilderImpl[TV] {
	bld.distance = angular.Distance[TV](vectorLength)
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV]) DotProductDistance(vectorLength int) *AnnoyIndexBuilderImpl[TV] {
	bld.distance = dotproduct.Distance[TV](vectorLength)
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV]) EuclideanDistance(vectorLength int) *AnnoyIndexBuilderImpl[TV] {
	bld.distance = euclidean.Distance[TV](vectorLength)
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV]) ManhattanDistance(vectorLength int) *AnnoyIndexBuilderImpl[TV] {
	bld.distance = manhattan.Distance[TV](vectorLength)
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV]) HammingDistance(vectorLength int) *AnnoyIndexBuilderImpl[TV] {
	bld.distance = hamming.Distance[TV](vectorLength)
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV]) UseMultiWorkerPolicy() *AnnoyIndexBuilderImpl[TV] {
	bld.buildPolicy = policy.MultiWorker()
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV]) SingleWorkerPolicy() *AnnoyIndexBuilderImpl[TV] {
	bld.buildPolicy = policy.SingleWorker()
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV]) MmapIndexAllocator() *AnnoyIndexBuilderImpl[TV] {
	bld.indexMemoryAllocator = memory.MmapIndexAllocator()
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV]) GCMemoryIndexAllocator() *AnnoyIndexBuilderImpl[TV] {
	bld.indexMemoryAllocator = memory.FileIndexMemoryAllocator()
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV]) UseSorter(sorter interfaces.Sorter[TV]) *AnnoyIndexBuilderImpl[TV] {
	bld.sorter = sorter
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV]) VerboseLogging() *AnnoyIndexBuilderImpl[TV] {
	bld.logVerbose = true
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV]) Build() interfaces.AnnoyIndex[TV] {

	if bld.buildPolicy == nil {
		bld.buildPolicy = policy.SingleWorker()
	}

	if bld.allocator == nil {
		bld.allocator = memory.GoGCIndexAllocator()
	}

	if bld.indexMemoryAllocator == nil {
		bld.indexMemoryAllocator = memory.MmapIndexAllocator()
	}

	if bld.random == nil {
		bld.random = random.NewKiss64Random(0)
	}

	return index.New(
		bld.random,
		bld.distance,
		bld.buildPolicy,
		bld.allocator,
		bld.indexMemoryAllocator,
		bld.sorter,
		bld.logVerbose,
		bld.allocHint,
	)
}
