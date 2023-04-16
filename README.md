# goannoy

GoAnnoy is an efficient Approximate Nearest Neighbors library for Go, optimized for memory usage and fast loading/saving to disk. This is a complete, standalone port that does not rely on cgo or other interop with C++ code. GoAnnoy is inspired by the Spotify's [Annoy](https://github.com/spotify/annoy) library.

## Key Features

* Memory-efficient nearest neighbor search
* Fast disk loading and saving
* Standalone Go implementation, no need for cgo or C++ dependencies
* Supports custom distance functions and indexing policies
* Pluggable memory, file allocators

## Getting started

```go
// Create a annoy index and configure it
idx := index.NewAnnoyIndexImpl[float32, uint32](
		vectorLength,
		random.NewKiss32Random(uint32(0)),
		angular.Distance[float32](vectorLength),
		policy.MultiWorker(),
		memory.IndexMemoryAllocator(),
		memory.MmapIndexAllocator(),
		false, /*verbose*/
		0,
	)

// Add some vectors and build the index
idx.AddItem(0, []float32{0, 0, 1})
idx.AddItem(1, []float32{0, 1, 0})
idx.AddItem(2, []float32{1, 0, 0})
idx.Build(10, -1)

ctx := idx.CreateContext()

// Now it is possible to search the index (in memory)
result, _ := idx.GetNnsByVector([]float32{3, 2, 1}, 3, -1, ctx)
assert.Equal(t, []uint32{2, 1, 0}, result)

// Save the index for later use (binary)
idx.Save("test.ann")

// Load it back at a later point in time and start searching.
//
// NOTE: It is possible to share the index with many processes.
idx.Load("test.ann")

// and more...
```

## Credits

This is a port of the Spotify https://github.com/spotify/annoy - all kudos goes to them! :)
