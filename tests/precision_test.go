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

	numItems := 10000 //1000000
	vectorLength := 40
	randomVectorContents := true
	multiplier := 2

	idx := index.NewAnnoyIndexImpl[float64, uint32](
		vectorLength,
		rnd,
		angular.Distance[float64, uint32](vectorLength),
		policy.Multi(),
		allocator,
		memory.MmapIndexAllocator(),
		false, /*verbose*/
	)

	defer idx.Close()

	vec_rnd := random.NewGoRandom()

	createVector := func() []float64 {
		vec := make([]float64, vectorLength)
		for z := 0; z < vectorLength; z++ {
			if randomVectorContents {
				vec[z] = vec_rnd.NormFloat64()
			} else {
				vec[z] = float64(z + 1)
			}
		}

		return vec
	}

	vectors := make([][]float64, numItems)
	for i := 0; i < numItems; i++ {
		v := createVector()
		vectors[i] = v
		idx.AddItem(i, v)
	}

	// output resulting metrics to file results.txt
	f, err := os.Create("results.txt")
	require.NoError(t, err)

	defer f.Close()

	f.WriteString(
		fmt.Sprintf(
			"numItems: %d, vectorLength: %d, multiplier: %d, randomVectorContents: %t\n",
			numItems, vectorLength, multiplier, randomVectorContents),
	)

	dur := utils.Measure(func() {
		idx.Build(multiplier*vectorLength, -1)
	})

	f.WriteString(fmt.Sprintf("build time: %d ms\n", dur.Milliseconds()))
	f.WriteString(fmt.Sprintln("Saving index ..."))

	defer os.Remove("test.ann")

	dur, err = utils.MeasureWithReturn(func() error {
		return idx.Save("test.ann")
	})

	require.NoError(t, err)

	f.WriteString(fmt.Sprintf("Saved in %d ms\n", dur.Milliseconds()))

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

	var limits []int
	for i := 1; i <= numItems; i *= 10 {
		limits = append(limits, i)
	}

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

		f.WriteString(fmt.Sprintf("finding nbs for %d\n", j))

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

	for _, limit := range limits {
		prec := prec_sum[limit] / float64(prec_n)
		time := time_sum[limit] / float64(prec_n)

		f.WriteString(
			fmt.Sprintf("limit = %d, precision = %f, time = %f ms\n", limit, prec, time),
		)
	}
}
