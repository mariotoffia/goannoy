package tests

import (
	"fmt"
	"testing"

	"github.com/mariotoffia/goannoy/distance/angular"
	"github.com/mariotoffia/goannoy/index"
	"github.com/mariotoffia/goannoy/policy"
	"github.com/mariotoffia/goannoy/random"
)

func TestPrecisionSingleThreaded(t *testing.T) {
	rnd := random.NewKiss32Random(uint32(0) /*default seed*/)
	allocator := index.NewArenaAllocator()

	defer allocator.Free()

	n := 100
	vectorLength := 1536

	index := index.NewAnnoyIndexImpl[float64, uint32](
		vectorLength,
		rnd,
		&angular.AngularDistanceImpl[float64, uint32]{},
		&policy.AnnoyIndexSingleThreadedBuildPolicy{},
		allocator,
	)

	vec_rnd := random.NewGoRandom()

	createVector := func() []float64 {
		vec := make([]float64, vectorLength)
		for z := 0; z < vectorLength; z++ {
			vec[z] = vec_rnd.NormFloat64()
		}

		return vec
	}

	for i := 0; i < n; i++ {
		index.AddItem(i, createVector())
	}

	fmt.Println("Building index num_trees = 2 * num_features ...")
	index.Build(2 * vectorLength)
	fmt.Println("Done building index")
}
