# goannoy

GoAnnoy is an efficient Approximate Nearest Neighbors library for Go, optimized for memory usage and fast loading/saving to disk. This is a complete, standalone port that does not rely on cgo or other interop with C++ code. GoAnnoy is a port of Spotify's [Annoy](https://github.com/spotify/annoy) library.

## Key Features

* Memory-efficient nearest neighbor search (using `unsafe` to handle unions, variable vector length and do continuous memory mapping)
* Fast disk loading and saving
* Standalone Go implementation, no need for cgo or C++ dependencies
* Supports custom distance functions and indexing policies (e.g. multi-threaded)
* Pluggable memory, file allocators

## Use Cases

* Approximate nearest neighbor search
* Recommendation systems
* Clustering
* Store of embeddings

## Getting started

```go
// Create a annoy index and configure it
idx := 	builder.Index().
		AngularDistance(1536 /*vectorLength*/).
		UseMultiWorkerPolicy().
		MmapIndexAllocator().
		Build()

// NOTE: If your'e adding huge amount of items to the index,
//       use the IndexNumHint(numIdx*numTrees) to pre-allocate and hence
//       it is much faster producing the index.

// Add some vectors and build the index
idx.AddItem(0, []float32{0, 0, 1})
idx.AddItem(1, []float32{0, 1, 0})
idx.AddItem(2, []float32{1, 0, 0})
idx.Build(10, -1)

ctx := idx.CreateContext()

// Now it is possible to search the index (in memory)
result, _ := idx.GetNnsByVector([]float32{3, 2, 1}, 3, -1, ctx)
assert.Equal(t, []uint32{2, 1, 0}, result)

// Save the index for later use
idx.Save("test.ann")

// Load it back at a later point in time and start searching.
idx.Load("test.ann")

// ...
```

## Precision Test Command Line Tool

Use the `go run cmd/precision/main.go` to test a few aspects of indexing and querying the vector index. It supports the following command line parameters:

```bash
Usage of precision:
  -cpu-profile
    	Enable CPU profiling
  -file
    	Write output to file results.txt (default to stdout)
  -items int
    	Number of items to create (default 1000)
  -keep
    	Keep the .ann file
  -length int
    	Vector length (default 40)
  -mem-profile
    	Enable memory profiling (go tool pprof /path/to/profile)
  -prec int
    	Number of items to test precision for (default 1000)
  -use-memory-index-allocator
    	Use memory index allocator (default is mmap)
  -verbose
    	Verbose output
```

For example, use the following:
```bash
go run cmd/precision/main.go -file -items 10000 -prec 1000
```
will generate *10_000* indexes and search the index. A _results.txt_ in the current directory is created with performance stats.

## Optional CPU Acceleration (Experimental)

> **Note:** CPU acceleration is experimental. The accelerated backends produce numerically equivalent results (within floating-point tolerance) but depend on build tags and, on `amd64`, an experimental Go feature. APIs and build requirements may change.

The default build uses the scalar Go dot-product implementation. To opt into hardware acceleration, add the `accelerate` build tag. The `GOARCH` target selects the backend automatically:

| Target | Backend | Build tag | Extra requirement |
|--------|---------|-----------|-------------------|
| `arm64` | NEON/AdvSIMD assembly | `accelerate` | — |
| `amd64` | Go 1.26 `simd/archsimd` intrinsics | `accelerate` | `GOEXPERIMENT=simd` |
| Any other | Scalar (no-op) | — | — |

Builds **without** the `accelerate` tag always use the portable scalar implementation regardless of architecture.

Examples:

```bash
# Default scalar build (all platforms)
go build ./...

# arm64 accelerated build (Apple Silicon, Graviton, etc.)
go build -tags accelerate ./...

# amd64 accelerated build (requires Go 1.26+ experimental SIMD support)
GOEXPERIMENT=simd go build -tags accelerate ./...
```

If you use the repository `Makefile`, note that it already exports `GOEXPERIMENT=arenas`. For accelerated `amd64` builds, use `GOEXPERIMENT=arenas,simd`.

## Benchmarking Scalar vs Accelerated Builds

The acceleration benchmarks include:

* primitive dot-product benchmarks in [`vector`](./vector)
* end-to-end indexing benchmarks
* end-to-end search benchmarks for `AngularDistance` and `DotProductDistance`

Run the current build's benchmark suite:

```bash
make bench
```

Run scalar-only output and save it to `bench-results/`:

```bash
make bench-scalar
```

Run accelerated benchmarks on Apple Silicon or other `arm64` machines:

```bash
make bench-accel-arm64
```

Run accelerated benchmarks on `amd64` machines:

```bash
make bench-accel-amd64
```

Run both scalar and accelerated benchmarks back-to-back and compare them:

```bash
make bench-compare
```

`bench-compare` saves both runs under `bench-results/`. If [`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) is installed, it is run automatically. Otherwise, compare the two output files manually.

## Credits

This is a port of Spotify https://github.com/spotify/annoy - all kudos goes to them! :)
