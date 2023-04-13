package tests

import (
	"bytes"
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

// https://github.com/erikbern/ann-benchmarks

func TestPrecision(t *testing.T) {
	numItems := uint32(1000000) //1000000
	vectorLength := uint32(40)
	randomVectorContents := true
	multiplier := uint32(2)
	verbose := false
	justGenerate := true
	keepAnnFile := true

	var buffer bytes.Buffer

	rnd := random.NewKiss32Random(uint32(0) /*default seed*/)
	allocator := memory.IndexMemoryAllocator()

	defer allocator.Free()

	idx := index.NewAnnoyIndexImpl[float32, uint32](
		vectorLength,
		rnd,
		angular.Distance[float32](vectorLength),
		policy.MultiWorker(),
		allocator,
		memory.MmapIndexAllocator(),
		verbose,
		numItems*multiplier,
	)

	defer idx.Close()

	vec_rnd := random.NewGoRandom()

	createVector := func() []float32 {
		vec := make([]float32, vectorLength)
		for z := uint32(0); z < vectorLength; z++ {
			if randomVectorContents {
				vec[z] = float32(vec_rnd.NormFloat64())
			} else {
				vec[z] = float32(z + 1)
			}
		}

		return vec
	}

	fmt.Fprintf(
		&buffer, "Create index: numItems: %d, vectorLength: %d, multiplier: %d, randomVectorContents: %t\n",
		numItems, vectorLength, multiplier, randomVectorContents,
	)

	vectors := make([][]float32, numItems)

	dur := utils.Measure(func() {
		for i := uint32(0); i < numItems; i++ {
			v := createVector()
			vectors[i] = v
			idx.AddItem(i, v)
		}
	})

	fmt.Fprintf(&buffer, "Index creation time: %d ms\n", dur.Milliseconds())

	fmt.Fprintf(
		&buffer, "numItems: %d, vectorLength: %d, multiplier: %d, randomVectorContents: %t\n",
		numItems, vectorLength, multiplier, randomVectorContents)

	dur = utils.Measure(func() {
		idx.Build(int(multiplier*vectorLength), -1)
	})

	fmt.Fprintf(&buffer, "Build time: %d ms\n", dur.Milliseconds())
	fmt.Fprintf(&buffer, "Saving index ...")

	defer func() {
		if !keepAnnFile {
			os.Remove("test.ann")
		}
	}()

	var err error

	dur, err = utils.MeasureWithReturn(func() error {
		return idx.Save("test.ann")
	})

	require.NoError(t, err)

	fmt.Fprintf(&buffer, "Saved in %d ms\n", dur.Milliseconds())

	defer func() {
		// output resulting metrics to file results.txt
		f, err := os.Create("results.txt")
		require.NoError(t, err)

		defer f.Close()
		f.WriteString(buffer.String())

	}()

	if justGenerate {
		return
	}
	for i := uint32(0); i < numItems; i++ {
		v := vectors[i]
		iv := idx.GetItemVector(i)

		// Compare vectors
		for j := uint32(0); j < vectorLength; j++ {
			if v[j] != iv[j] {
				t.Fatalf("Vector mismatch at index %d, %f != %f", j, v[j], iv[j])
			}
		}
	}

	var limits []int
	for i := 1; i <= int(numItems); i *= 10 {
		limits = append(limits, i)
		if i == 100 {
			limits = append(limits, 500)
		}
	}

	numReturn := 10
	prec_n := 1000
	prec_sum := make(map[int]float64)
	time_sum := make(map[int]float64)
	var closest []uint32

	// init precision and timers map
	for _, limit := range limits {
		prec_sum[limit] = 0.0
		time_sum[limit] = 0.0
	}

	// doing the work
	for i := 0; i < prec_n; i++ {
		// select a random node
		j := rnd.NextIndex(uint32(numItems))

		fmt.Fprintf(&buffer, "finding nbs for %d\n", j)

		// getting the K closest
		closest, _ = idx.GetNnsByItem(j, numReturn, int(numItems))

		for _, limit := range limits {

			dur, topList := utils.MeasureWithReturn(func() []uint32 {
				c, _ := idx.GetNnsByItem(j, limit, -1)
				return c
			})

			time_sum[limit] += float64(dur.Microseconds())

			// intersecting results
			found := len(util.Intersection(closest, topList))
			hitRate := float64(found) / float64(numReturn)
			prec_sum[limit] += hitRate
		}
	}

	for _, limit := range limits {
		prec := prec_sum[limit] / float64(prec_n)
		time := time_sum[limit] / float64(prec_n)

		if time >= 1000 {
			fmt.Fprintf(
				&buffer, "limit = %d, precision = %f, time = %f ms\n", limit, prec, time/1000,
			)
		} else {
			fmt.Fprintf(
				&buffer, "limit = %d, precision = %f, time = %f us\n", limit, prec, time,
			)
		}
	}
}
