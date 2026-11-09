package main

import (
	"flag"
	"fmt"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/angular"
	"io"
	"math"
	"math/rand"
	"os"
	"slices"
	"time"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("precision", flag.ContinueOnError)
	flags.SetOutput(out)
	items := flags.Int("items", 1000, "number of vectors")
	dimension := flags.Int("length", 40, "vector dimension")
	queries := flags.Int("prec", 100, "number of queries")
	trees := flags.Int("trees", 10, "number of trees")
	workers := flags.Int("workers", 0, "build workers (0 = available CPUs)")
	searchK := flags.Int("search-k", -1, "candidate budget (-1 = automatic)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *items < 1 || *dimension < 1 || *queries < 1 {
		return fmt.Errorf("items, length and prec must be positive")
	}
	builder, err := angular.NewBuilder(*dimension)
	if err != nil {
		return err
	}
	builder.Trees(*trees).Workers(*workers).Seed(42).Capacity(*items)
	rng := rand.New(rand.NewSource(42))
	vectors := make([][]float32, *items)
	for id := range vectors {
		vectors[id] = make([]float32, *dimension)
		for j := range vectors[id] {
			vectors[id][j] = float32(rng.NormFloat64())
		}
		if err := builder.Add(goannoy.ItemID(id), vectors[id]); err != nil {
			return err
		}
	}
	start := time.Now()
	index, err := builder.Build(nil)
	if err != nil {
		return err
	}
	defer index.Close()
	buildTime := time.Since(start)
	searcher := index.NewSearcher()
	limit := min(10, *items)
	dst := make([]goannoy.Result, limit)
	hits := 0
	queryTime := time.Duration(0)
	for n := 0; n < *queries; n++ {
		q := vectors[rng.Intn(*items)]
		exact := make([]goannoy.Result, *items)
		for id, v := range vectors {
			exact[id] = goannoy.Result{ID: goannoy.ItemID(id), Distance: referenceAngular(q, v)}
		}
		slices.SortFunc(exact, func(a, b goannoy.Result) int {
			if a.Distance < b.Distance || a.Distance == b.Distance && a.ID < b.ID {
				return -1
			}
			if a.Distance == b.Distance && a.ID == b.ID {
				return 0
			}
			return 1
		})
		start = time.Now()
		count, err := searcher.SearchInto(q, dst, goannoy.SearchK(*searchK))
		if err != nil {
			return err
		}
		queryTime += time.Since(start)
		for _, r := range dst[:count] {
			for _, want := range exact[:limit] {
				if r.ID == want.ID {
					hits++
					break
				}
			}
		}
	}
	_, err = fmt.Fprintf(out, "items=%d dimension=%d trees=%d build=%s recall@%d=%.4f query=%s\n",
		*items, *dimension, index.Trees(), buildTime, limit, float64(hits)/float64(*queries*limit), queryTime/time.Duration(*queries))
	return err
}
func referenceAngular(a, b []float32) float32 {
	var aa, bb, ab float64
	for j, x := range a {
		y := float64(b[j])
		v := float64(x)
		aa += v * v
		bb += y * y
		ab += v * y
	}
	if aa*bb == 0 {
		return float32(math.Sqrt(2))
	}
	return float32(math.Sqrt(max(0, 2-2*ab/math.Sqrt(aa*bb))))
}
