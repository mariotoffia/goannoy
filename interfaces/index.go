package interfaces

import (
	"github.com/mariotoffia/goannoy/distance"
	"github.com/mariotoffia/goannoy/random"
)

type AnnoyIndex[
	TIdx int,
	TV distance.VectorType,
	TRandType random.RandomTypes,
	TRand random.Random[TRandType]] interface {
}

type AnnoyIndexBuilder interface {
	ThreadBuild(q int, threadIdx int, threadedBuildPolicy AnnoyIndexBuildPolicy)
}
