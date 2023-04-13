package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/mariotoffia/goannoy/distance/angular"
	"github.com/mariotoffia/goannoy/index"
	"github.com/mariotoffia/goannoy/index/memory"
	"github.com/mariotoffia/goannoy/index/policy"
	"github.com/mariotoffia/goannoy/random"
	"github.com/mariotoffia/goannoy/utils"
)

func main() {
	numItems := 1000
	vectorLength := 40
	randomVectorContents := true
	multiplier := uint32(2)
	verbose := false
	justGenerate := false
	keepAnnFile := false
	toFile := false
	numReturn := 10
	prec_n := 1000

	flag.BoolVar(&toFile, "file", false, "Write output to file results.txt")
	flag.BoolVar(&keepAnnFile, "keep", false, "Keep the .ann file")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.IntVar(&numItems, "items", 1000, "Number of items to create")
	flag.IntVar(&vectorLength, "length", 40, "Vector length")

	flag.Parse()

	var buffer io.Writer

	if toFile {
		buffer = &bytes.Buffer{}
	} else {
		buffer = os.Stdout
	}

	rnd := random.NewKiss32Random(uint32(0) /*default seed*/)

	idx := index.NewAnnoyIndexImpl[float32, uint32](
		uint32(vectorLength),
		rnd,
		angular.Distance[float32](uint32(vectorLength)),
		policy.MultiWorker(),
		memory.IndexMemoryAllocator(),
		memory.MmapIndexAllocator(),
		verbose,
		uint32(numItems)*multiplier, /*alloc hint for faster build*/
	)

	defer idx.Close()

	vec_rnd := random.NewGoRandom()

	createVector := func() []float32 {
		vec := make([]float32, vectorLength)
		for z := uint32(0); z < uint32(vectorLength); z++ {
			if randomVectorContents {
				vec[z] = float32(vec_rnd.NormFloat64())
			} else {
				vec[z] = float32(z + 1)
			}
		}

		return vec
	}

	fmt.Fprintf(
		buffer, "Create index: numItems: %d, vectorLength: %d, multiplier: %d, randomVectorContents: %t\n",
		numItems, vectorLength, multiplier, randomVectorContents,
	)

	vectors := make([][]float32, numItems)

	dur := utils.Measure(func() {
		for i := 0; i < numItems; i++ {
			v := createVector()
			vectors[i] = v
			idx.AddItem(uint32(i), v)
		}
	})

	fmt.Fprintf(buffer, "Index creation time: %d ms\n", dur.Milliseconds())

	fmt.Fprintf(
		buffer, "numItems: %d, vectorLength: %d, multiplier: %d, randomVectorContents: %t\n",
		numItems, vectorLength, multiplier, randomVectorContents)

	dur = utils.Measure(func() {
		idx.Build(int(multiplier*uint32(vectorLength)), -1)
	})

	fmt.Fprintf(buffer, "Build time: %d ms\n", dur.Milliseconds())
	fmt.Fprintf(buffer, "Saving index ...\n")

	defer func() {
		if !keepAnnFile {
			os.Remove("test.ann")
		}
	}()

	var err error

	dur, err = utils.MeasureWithReturn(func() error {
		return idx.Save("test.ann")
	})

	if err != nil {
		panic(fmt.Sprintf("Error creating file: %s", err.Error()))
	}

	fmt.Fprintf(buffer, "Saved in %d ms\n", dur.Milliseconds())

	defer func() {
		if !toFile {
			return
		}
		// output resulting metrics to file results.txt
		f, err := os.Create("results.txt")
		if err != nil {
			panic(fmt.Sprintf("Error creating file: %s", err.Error()))
		}

		defer f.Close()
		f.WriteString(buffer.(*bytes.Buffer).String())

	}()

	if justGenerate {
		return
	}
	for i := 0; i < numItems; i++ {
		v := vectors[i]
		iv := idx.GetItemVector(uint32(i))

		// Compare vectors
		for j := uint32(0); j < uint32(vectorLength); j++ {
			if v[j] != iv[j] {
				panic(fmt.Sprintf("Vector mismatch at index %d, %f != %f", j, v[j], iv[j]))
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

		fmt.Fprintf(buffer, "finding nbs for %d\n", j)

		// getting the K closest
		closest, _ = idx.GetNnsByItem(j, numReturn, int(numItems))

		for _, limit := range limits {

			dur, topList := utils.MeasureWithReturn(func() []uint32 {
				c, _ := idx.GetNnsByItem(j, limit, -1)
				return c
			})

			time_sum[limit] += float64(dur.Microseconds())

			// intersecting results
			found := len(utils.Intersection(closest, topList))
			hitRate := float64(found) / float64(numReturn)
			prec_sum[limit] += hitRate
		}
	}

	for _, limit := range limits {
		prec := prec_sum[limit] / float64(prec_n)
		time := time_sum[limit] / float64(prec_n)

		if time >= 1000 {
			fmt.Fprintf(
				buffer, "limit = %d, precision = %f, time = %f ms\n", limit, prec, time/1000,
			)
		} else {
			fmt.Fprintf(
				buffer, "limit = %d, precision = %f, time = %f us\n", limit, prec, time,
			)
		}
	}
}
