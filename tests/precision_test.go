package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/mariotoffia/goannoy/distance/angular"
	"github.com/mariotoffia/goannoy/index"
	"github.com/mariotoffia/goannoy/index/memory"
	"github.com/mariotoffia/goannoy/index/policy"
	"github.com/mariotoffia/goannoy/random"
	"github.com/mariotoffia/goannoy/tests/utils"
	util "github.com/mariotoffia/goannoy/utils"
	"github.com/stretchr/testify/require"
)

func TestPrecision(t *testing.T) {
	rnd := random.NewKiss32Random(uint32(0) /*default seed*/)
	allocator := memory.NewBuildIndexMemoryArenaAllocator()

	defer allocator.Free()

	numItems := 100
	vectorLength := 40

	idx := index.NewAnnoyIndexImpl[float64, uint32](
		vectorLength,
		rnd,
		&angular.AngularDistanceImpl[float64, uint32]{},
		policy.Multi(),
		allocator,
		memory.MmapIndexAllocator(),
	)

	defer idx.Close()

	vec_rnd := random.NewGoRandom()

	createVector := func() []float64 {
		vec := make([]float64, vectorLength)
		for z := 0; z < vectorLength; z++ {
			vec[z] = vec_rnd.NormFloat64()
		}

		return vec
	}

	vectors := make([][]float64, numItems)
	for i := 0; i < numItems; i++ {
		v := createVector()
		vectors[i] = v
		idx.AddItem(i, v)
	}

	fmt.Printf("Building index num_trees = 2 * vectorLength (%d) ...\n", 2*vectorLength)
	idx.Build(2*vectorLength, -1)
	fmt.Println("Done building index")

	fmt.Println("Saving index ...")

	defer os.Remove("test.ann")

	dur, err := utils.MeasureWithReturn(func() error {
		return idx.Save("test.ann")
	})

	require.NoError(t, err)

	fmt.Printf("Done in %d ms\n", dur.Milliseconds())

	for i := 0; i < numItems; i++ {
		v := vectors[i]
		iv := idx.GetItemVector(i)

		// Compare vectors
		for j := 0; j < vectorLength; j++ {
			if v[j] != iv[j] {
				t.Fatalf("Vector mismatch at index %d, %f != %f", j, v[j], iv[j])
			}
		}
	}

	limits := []int{1, 10, 100, 1000, 10000, 100000, 1000000}
	numReturn := 10
	prec_n := 1000
	prec_sum := make(map[int]float64)
	time_sum := make(map[int]float64)
	var closest []int

	// init precision and timers map
	for _, limit := range limits {
		prec_sum[limit] = 0.0
		time_sum[limit] = 0.0
	}

	// doing the work
	for i := 0; i < prec_n; i++ {
		// select a random node
		j := int(rnd.NextIndex(uint32(numItems)))

		fmt.Println("finding nbs for ", j)

		// getting the K closest
		closest, _ = idx.GetNnsByItem(j, numReturn, numItems)

		for _, limit := range limits {

			dur, topList := utils.MeasureWithReturn(func() []int {
				c, _ := idx.GetNnsByItem(j, limit, -1)
				return c
			})

			time_sum[limit] += float64(dur.Milliseconds())

			// intersecting results
			found := len(util.Intersection(closest, topList))
			hitRate := float64(found) / float64(numReturn)
			prec_sum[limit] += hitRate
		}
	}

	// printing resulting metrics
	for _, limit := range limits {
		prec := prec_sum[limit] / float64(prec_n)
		time := time_sum[limit] / float64(prec_n)
		fmt.Printf("limit = %d, precision = %f, time = %f ms\n", limit, prec, time)
	}
}
