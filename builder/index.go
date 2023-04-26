package builder

import (
	"github.com/mariotoffia/goannoy/distance/angular"
	"github.com/mariotoffia/goannoy/index"
	"github.com/mariotoffia/goannoy/index/memory"
	"github.com/mariotoffia/goannoy/index/policy"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/random"
)

type AnnoyIndexBuilderImpl[TV interfaces.VectorType, TIX interfaces.IndexTypes] struct {
	allocHint            TIX
	random               interfaces.Random[TIX]
	distance             interfaces.Distance[TV, TIX]
	buildPolicy          interfaces.AnnoyIndexBuildPolicy
	allocator            interfaces.BuildIndexAllocator
	indexMemoryAllocator interfaces.IndexAllocator
	sorter               interfaces.Sorter[TV, TIX]
	logVerbose           bool
}

// Index creates a new `AnnoyIndexBuilderImpl` instance.
func Index[TV interfaces.VectorType, TIX interfaces.IndexTypes]() *AnnoyIndexBuilderImpl[TV, TIX] {
	return &AnnoyIndexBuilderImpl[TV, TIX]{}
}

func (bld *AnnoyIndexBuilderImpl[TV, TIX]) Random(rnd interfaces.Random[TIX]) *AnnoyIndexBuilderImpl[TV, TIX] {
	bld.random = rnd
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV, TIX]) IndexNumHint(allocHint int) *AnnoyIndexBuilderImpl[TV, TIX] {
	if allocHint <= 0 {
		return bld
	}

	bld.allocHint = TIX(allocHint)
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV, TIX]) AngularDistance(vectorLength int) *AnnoyIndexBuilderImpl[TV, TIX] {
	bld.distance = angular.Distance[TV](TIX(vectorLength))
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV, TIX]) UseMultiWorkerPolicy() *AnnoyIndexBuilderImpl[TV, TIX] {
	bld.buildPolicy = policy.MultiWorker()
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV, TIX]) SingleWorkerPolicy() *AnnoyIndexBuilderImpl[TV, TIX] {
	bld.buildPolicy = policy.SingleWorker()
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV, TIX]) MmapIndexAllocator() *AnnoyIndexBuilderImpl[TV, TIX] {
	bld.indexMemoryAllocator = memory.MmapIndexAllocator()
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV, TIX]) GCMemoryIndexAllocator() *AnnoyIndexBuilderImpl[TV, TIX] {
	bld.indexMemoryAllocator = memory.FileIndexMemoryAllocator()
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV, TIX]) UseSorter(sorter interfaces.Sorter[TV, TIX]) *AnnoyIndexBuilderImpl[TV, TIX] {
	bld.sorter = sorter
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV, TIX]) VerboseLogging() *AnnoyIndexBuilderImpl[TV, TIX] {
	bld.logVerbose = true
	return bld
}

func (bld *AnnoyIndexBuilderImpl[TV, TIX]) Build() interfaces.AnnoyIndex[TV, TIX] {

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
		var t TIX

		switch any(t).(type) {
		case uint32:
			k := random.NewKiss32Random(uint32(0))
			bld.random = any(k).(interfaces.Random[TIX]) // Ugly hack to get around type system
		case uint64:
			// TODO: Need to completely refactor random!!
		}
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
