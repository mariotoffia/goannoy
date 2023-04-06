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

	n := 10000
	vectorLength := 40

	idx := index.NewAnnoyIndexImpl[float64, uint32](
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

	vectors := make([][]float64, n)
	for i := 0; i < n; i++ {
		v := createVector()
		vectors[i] = v
		idx.AddItem(i, v)
	}

	fmt.Println("Building index num_trees = 2 * num_features ...")
	idx.Build(2 * vectorLength)
	fmt.Println("Done building index")

	fmt.Println("Saving index ...")
	idx.Save("test.ann")
	fmt.Println("Done")

	for i := 0; i < n; i++ {
		v := vectors[i]
		iv := idx.GetItemVector(i)

		// Compare vectors
		for j := 0; j < vectorLength; j++ {
			if v[j] != iv[j] {
				t.Errorf("Vector mismatch at index %d, %f != %f", j, v[j], iv[j])
			}
		}
	}
}
