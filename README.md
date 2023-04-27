# goannoy

GoAnnoy is an efficient Approximate Nearest Neighbors library for Go, optimized for memory usage and fast loading/saving to disk. This is a complete, standalone port that does not rely on cgo or other interop with C++ code. GoAnnoy is a port of Spotify's [Annoy](https://github.com/spotify/annoy) library.

## Key Features

* Memory-efficient nearest neighbor search (using `unsafe` to handle unions, variable vector length and do continuous memory mapping)
* Fast disk loading and saving
* Standalone Go implementation, no need for cgo or C++ dependencies
* Supports custom distance functions and indexing policies (e.g. multi-threaded)
* Pluggable memory, file allocators

## Getting started

```go
// Create a annoy index and configure it
idx := 	builder.Index[float32, uint32]().
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

## Credits

This is a port of Spotify https://github.com/spotify/annoy - all kudos goes to them! :)

