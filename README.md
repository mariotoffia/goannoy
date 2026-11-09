# GoAnnoy

> **Breaking API changes in v0.1.0**
>
> This release replaces the earlier GoAnnoy API. If you are upgrading existing
> code, ask your coding agent to use the [GoAnnoy upgrade skill](skills/goannoy-v01-upgrade/SKILL.md).
> It maps old builder and search calls to the new API and identifies work that
> needs manual review, including legacy Hamming index rebuilds.

Fast similarity search for Go, based on [Spotify Annoy](https://github.com/spotify/annoy).
No server, no cgo, and no external runtime dependencies.

GoAnnoy finds items with similar **vectors**: lists of numbers describing things
like documents, products, images, or songs. Your application supplies the vectors;
GoAnnoy builds a searchable index and returns nearby item IDs and distances.
Keep the original content and metadata in your own storage.

Use it for recommendations, related-item lookup, or searching embeddings you
already have. It is a library inside your application—not a database server or
an embedding model.

[Quick start](#quick-start) · [API reference](https://pkg.go.dev/github.com/mariotoffia/goannoy) · [Release notes](https://github.com/mariotoffia/goannoy/releases/tag/v0.1.0) · [Upgrade guide](skills/goannoy-v01-upgrade/SKILL.md)

## Install

Requires **Go 1.26 or later**. Inside a Go module:

```sh
go get github.com/mariotoffia/goannoy@v0.1.0
```

**v0.1.0 is the first release.** The API may change before v1.0. If you used an
older unreleased GoAnnoy API, follow the upgrade guide rather than copying its
old builder or search-context calls.

## Quick start

Build a three-dimensional angular index, then find the three nearest items:

```go
package main

import (
	"context"
	"fmt"
	"log"

	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/angular"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	builder, err := angular.NewBuilder(3)
	if err != nil {
		return err
	}
	builder.Trees(10).Workers(1).Seed(42)

	vectors := [][]float32{
		{0, 0, 1},
		{0, 1, 0},
		{1, 0, 0},
	}
	for id, vector := range vectors {
		if err := builder.Add(goannoy.ItemID(id), vector); err != nil {
			return err
		}
	}

	index, err := builder.Build(context.Background())
	if err != nil {
		return err
	}
	defer index.Close()

	searcher := index.NewSearcher()
	query := []float32{3, 2, 1}
	results, err := searcher.Search(query, 3, goannoy.AutoSearchK)
	if err != nil {
		return err
	}

	for _, result := range results {
		fmt.Printf("id=%d distance=%.3f\n", result.ID, result.Distance)
	}
	return nil
}
```

```text
id=2 distance=0.630
id=1 distance=0.965
id=0 distance=1.211
```

A built index is an immutable snapshot. To change the dataset, add or replace
items in the builder and build a new snapshot; existing indexes stay unchanged.

## Choose a metric

Import the package that matches your data or embedding model:

| Package | Compares | Input | Best result |
| --- | --- | --- | --- |
| [`angular`](https://pkg.go.dev/github.com/mariotoffia/goannoy/angular) | Vector direction | `[]float32` | Smallest distance |
| [`dot`](https://pkg.go.dev/github.com/mariotoffia/goannoy/dot) | Dot-product similarity | `[]float32` | Largest score |
| [`euclidean`](https://pkg.go.dev/github.com/mariotoffia/goannoy/euclidean) | Straight-line distance | `[]float32` | Smallest distance |
| [`manhattan`](https://pkg.go.dev/github.com/mariotoffia/goannoy/manhattan) | Sum of absolute coordinate differences | `[]float32` | Smallest distance |
| [`hamming`](https://pkg.go.dev/github.com/mariotoffia/goannoy/hamming) | Differing bits | Packed `[]uint64` | Smallest distance |

Angular uses Annoy's `sqrt(2 - 2*cosine)` distance, not a cosine-similarity score.
All metrics use ascending item ID to break equal-distance/score ties.

### Speed and accuracy

Search is **approximate**: it can skip potential matches to answer faster.
It does not guarantee the exact closest items on every query.

- More **trees** generally improve recall, at the cost of build time and storage.
- A larger **search budget** generally improves recall, at the cost of query time.
- `goannoy.AutoSearchK` uses requested neighbors × tree count. Other budgets must
  be positive.

Check recall with representative queries from your own dataset. The included
precision tool can help exercise the build/search workflow.

## Reuse buffers for repeated searches

`Search` allocates a result slice for convenience. Inside an error-returning
function, use `SearchInto` to reuse a buffer with an existing `index`:

```go
searcher := index.NewSearcher()
results := make([]goannoy.Result, 10)
query := []float32{3, 2, 1}

n, err := searcher.SearchInto(query, results, goannoy.AutoSearchK)
if err != nil {
	return err
}

for _, result := range results[:n] {
	fmt.Printf("id=%d distance=%.3f\n", result.ID, result.Distance)
}
```

`len(results)` is the requested result count; only `results[:n]` is populated.
Copy results you need to retain before reusing the buffer. After warm-up,
`SearchInto` allocates zero heap objects while its scratch capacity is sufficient.
Use `SearchByIDInto` to search with an existing item's stored vector.

### Concurrency and ownership

- Share an **index** between goroutines; give each concurrent caller its own
  **searcher and result buffer**. Builders and searchers are not concurrency-safe.
- `Add` copies its input. `Vector` returns a copy; `VectorInto` copies into your
  buffer. Neither exposes writable mapped storage. `Distance` compares stored items.
- `Close` is idempotent and waits for active operations. Later searches return
  `ErrClosed`.
- Construct handles with `NewBuilder`, `Build`, or `Open`; their zero values are
  not usable indexes, builders, or searchers.

See [concurrency and ownership](docs/concurrency.md) for details.

## Save and reopen an index

With the populated `builder` from the quick start:

```go
saved, err := builder.BuildTo(context.Background(), "vectors.ann")
if err != nil {
	return err
}
if err := saved.Close(); err != nil {
	return err
}

loaded, err := angular.Open("vectors.ann", 3)
if err != nil {
	return err
}
defer loaded.Close()
```

`BuildTo` opens and validates a completed temporary index before publishing it.
Existing open indexes keep their old snapshot. File replacement and durability
follow the platform's rename/sync guarantees; a directory-sync error can occur
after the complete new file has committed.

For file-backed tree construction, use `builder.OnDisk(path).Build(ctx)`.
OnDisk construction is serial and still keeps input vectors in memory.
Supported Unix targets use read-only mmap when opening files; other platforms
use owned-memory storage. `Open` also accepts the optional `goannoy.Prefault` flag.

## Configure the builder

All five metric builders expose the same small configuration surface:

| Method | Purpose |
| --- | --- |
| `Trees(n)` | Number of trees to build |
| `Workers(n)` | Build workers; explicit zero requests GOMAXPROCS |
| `Seed(seed)` | Reproducible random stream |
| `Capacity(n)` | Reserve input capacity to reduce growth |
| `Profile(profile)` | Balanced, Throughput, or LowMemory worker defaults |
| `OnDisk(path)` | File-backed, serial tree construction |

Defaults are ten trees and the Balanced profile. Without an explicit worker
setting, Balanced uses at most four workers, Throughput uses GOMAXPROCS, and
LowMemory uses one. Explicit `Workers` overrides the profile.

Invalid configuration is sticky: inspect `builder.Err()` or handle the next
`Add`/`Build` error. An invalid data vector does not poison the builder.
Nil build contexts are treated as `context.Background()`; cancellation preserves
builder input and existing index snapshots.

For the pinned serial C++ build topology, use `Workers(1)` with the same seed.
Parallel Go builds are repeatable for fixed settings, but are not byte-identical
to C++ parallel builds.

## Input and file compatibility

- Use nonnegative `goannoy.ItemID` values below `math.MaxInt32`. Prefer dense IDs:
  storage and `Items()` use maxID+1, not the number of inserted sparse items.
  Adding an existing ID replaces its vector; missing IDs are not valid targets.
- Float vectors require the configured length, finite components, and a finite
  float32 squared norm. Derived non-finite distances return `ErrInvalidVector`.
  Scale unusually large vectors; upstream angular rounding behavior is retained.
- Hamming takes a **bit dimension**, not a word count. A 130-bit vector needs
  three uint64 words, with only the low two bits usable in the final word.
  Unused tail bits must be zero. The maximum dimension is 2²⁴ bits, keeping all
  returned counts exactly representable in float32.
- Supported Spotify Annoy files use little-endian int32 IDs and float32 vectors,
  or packed uint64 Hamming words. Files are **headerless**: retain the metric,
  dimension, and layout externally. Validation cannot detect every wrong-layout
  combination. Do not modify or truncate files while they are mapped.
- Independent C++ fixtures and cross-language tests cover every metric. Defined
  serial fields match; C++-uninitialized split bytes are deliberately zeroed.
  Old GoAnnoy float-per-bit Hamming files must be rebuilt from source vectors.

Use `errors.Is` with the root package's sentinel errors for invalid input,
corrupt files, unsupported platforms, and closed indexes. See the
[upgrade skill](skills/goannoy-v01-upgrade/SKILL.md) for earlier API shapes and
[Hamming migration](skills/goannoy-v01-upgrade/references/hamming.md) for legacy files.

## Performance

Recorded angular workload: 10,000 × 128-dimensional vectors, ten trees, one
build worker, ten neighbors, Apple M4 Pro, Go 1.27.3. Figures are rounded;
the comparison is with the old **unreleased GoAnnoy branch**, not Spotify C++.

| Measurement | Before | v0.1.0 |
| --- | ---: | ---: |
| Scalar search latency | 36–46 µs | ~22 µs |
| ARM64 NEON search latency | 22–24 µs | ~10 µs |
| Warmed search allocations | ~100 | 0 |
| Heap bytes allocated per build | ~87 MB | ~26 MB |
| Minimal stripped angular executable | ~1.9 MB | ~1.7 MB |

Allocated bytes are **not peak RAM**. Reusable buffers, mapped files, and the
OS page cache still consume memory. See the [benchmark report](docs/release/v0.1.0-benchmarks.md)
for exact samples, timing variation, and reproduction commands.

Stable defaults use ARM64 NEON or AMD64 SSE2, with a scalar fallback elsewhere.
Use `-tags goannoy_scalar` to force scalar kernels. Index construction uses scalar
math for determinism. Experimental AMD64 Go SIMD requires both `GOEXPERIMENT=simd`
and `-tags goannoy_experimental_simd`, pinned to Go 1.27.3; ordinary use needs neither.

## Development

```sh
make setup build
make test
make test-race test-checkptr vet

go test -tags goannoy_scalar ./...
ANNOY_SOURCE=/path/to/pinned/annoy make compat

go run ./cmd/precision -items 10000 -length 128 -prec 100 -trees 10
```

`make test` enforces at least 90% core/public and 85% overall runtime coverage.
Ordinary tests use committed independent fixtures and need no C++ compiler.
The explicit [C++ compatibility gate](compat/README.md) requires the pinned
source/compiler and fails if unavailable. Generate traversal code with
`go generate ./internal/engine`; run `make bench` for benchmark workloads.

Further reading: [release evidence](docs/release/v0.1.0-checklist.md),
[design](docs/superpowers/specs/2026-11-09-goannoy-v0.1.0-design.md),
[implementation plan](docs/superpowers/plans/2026-11-09-goannoy-v0.1.0.md), and
[decisions and trade-offs](docs/release/v0.1.0-decisions.md).
