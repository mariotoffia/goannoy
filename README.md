# GoAnnoy

A Go-native port of [Spotify Annoy](https://github.com/spotify/annoy): approximate nearest-neighbor forests with standard Annoy index files, immutable concurrent indexes, and reusable search scratch. No cgo or external runtime dependencies. Go 1.26 or later.

The target is **v0.1.0**, an intentionally breaking pre-1.0 release. Earlier APIs are replaced, not retained behind adapters. Use the [agent upgrade skill](skills/goannoy-v01-upgrade/SKILL.md) when migrating an older consumer.

## Build and search

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
    b, err := angular.NewBuilder(3)
    if err != nil { log.Fatal(err) }
    b.Trees(10).Workers(1).Seed(42).Capacity(3)
    for id, vector := range [][]float32{{0, 0, 1}, {0, 1, 0}, {1, 0, 0}} {
        if err := b.Add(goannoy.ItemID(id), vector); err != nil { log.Fatal(err) }
    }
    idx, err := b.Build(context.Background())
    if err != nil { log.Fatal(err) }
    defer idx.Close()

    searcher := idx.NewSearcher()
    dst := make([]goannoy.Result, 3)
    n, err := searcher.SearchInto([]float32{3, 2, 1}, dst, goannoy.AutoSearchK)
    if err != nil { log.Fatal(err) }
    fmt.Println(dst[:n]) // IDs: 2, 1, 0
}
```

Choose the metric by importing its concrete package:

| Package | Vector | Returned value / order |
| --- | --- | --- |
| `angular` | `[]float32` | sqrt(2−2cosine), ascending |
| `dot` | `[]float32` | dot-product similarity, descending |
| `euclidean` | `[]float32` | L2 distance, ascending |
| `manhattan` | `[]float32` | L1 distance, ascending |
| `hamming` | `[]uint64` | differing-bit count, ascending |

Zero angular vectors follow Annoy's distance semantics. Ties use ascending item ID. Search is approximate: increasing the tree count or search budget generally improves recall. `AutoSearchK` (-1) uses requested neighbors × trees; other budgets must be positive. `SearchInto` uses **len(dst)** as the requested neighbor count and returns the populated prefix length. `Search` is an allocating convenience wrapper. `SearchByIDInto` searches an existing item's vector.

## Configuration and ownership

All metrics expose the same small fluent builder surface: `Trees`, `Workers`, `Seed`, `Capacity`, `Profile`, and `OnDisk`. Invalid configuration is sticky and available from `Err()` or the next `Add`/`Build`. Invalid input data returns an error without poisoning the builder.

Defaults are ten trees, the upstream Kiss64 seed `1234567890987654321`, and the Balanced profile. Balanced caps automatic workers at four, Throughput uses GOMAXPROCS, and LowMemory uses one. Explicit `Workers(n)` overrides the profile; `Workers(0)` requests GOMAXPROCS. For serial C++-compatible build topology, use `Workers(1)` with the same explicit seed. Parallel builds use deterministic independent tree streams: repeatability requires the same settings/worker count, not byte identity with C++ threading.

`Add` copies vectors. Adding an existing ID replaces it. IDs must satisfy `0 <= id < math.MaxInt32`; avoid huge sparse IDs because `Items()` and storage use maxID+1. Missing IDs are not valid search/vector targets. Builders may produce multiple independent immutable snapshots.

Share an index across goroutines; give each concurrent caller its own searcher and result buffer. Builders and searchers are not concurrency-safe. Warmed `SearchInto` allocates zero heap objects while its scratch capacity is sufficient. `Vector` and `VectorInto` copy data; they do not expose mapped memory. `Close` is idempotent and waits for active operations; later searches return `ErrClosed`. See [concurrency and profiles](docs/concurrency.md).

## Files and compatibility

Use `b.BuildTo(ctx, path)` to build and save, or `b.OnDisk(path).Build(ctx)` to file-back tree-node construction. OnDisk construction is serial and retains input vectors in memory. Output is written to a sibling temporary file, synced, and replaced; existing open indexes retain their old snapshot. Unix directory sync is included. Replacement atomicity follows the platform's rename guarantees; a directory-sync error may mean the complete new file is already committed.

```go
idx, err := angular.Open("vectors.ann", 128, goannoy.Prefault)
// Handle err and Close the index when finished.
```

Standard Annoy files are **headerless**: keep the metric, dimension, scalar layout, and provenance alongside each file. `Open` checks structural safety but cannot identify every wrong metric/dimension combination. This implementation targets little-endian, signed-int32/float32 standard files; Hamming words are uint64. Other layouts are unsupported. Unix targets use read-only mmap; other platforms use an owned-memory fallback.

Pinned C++ fixtures and Go-to-C++ cross-loading cover all five metrics. Serial builds match meaningful C++ fields; C++-uninitialized dot/Hamming split padding is intentionally zeroed, not reproduced. Search/build semantics match Annoy while internal Go storage and interfaces may differ. Files must not be truncated or modified externally while mapped.

Hamming takes the **bit dimension**, not word count. A 130-bit vector is three uint64 words; only the low two bits of the final word may be used. Bit j is bit j%64 of word j/64. Old GoAnnoy float-per-bit Hamming files are incompatible and require rebuilding from original vectors. See [Hamming migration](skills/goannoy-v01-upgrade/references/hamming.md).

Hamming dimensions are limited to **2²⁴ bits**, the exact-integer range of the
float32 result contract. Float metrics require finite components and a finite
float32 sum of squared components. A derived distance overflow returns
`ErrInvalidVector`, not NaN/Inf results. Scale large vectors appropriately;
finite-but-overflowing arithmetic is outside the supported numeric domain.
Within that domain upstream float32 rounding behavior is retained, including
angular norm-product overflow for unusually large magnitudes.

## Errors and extension

Use `errors.Is` with the root package's sentinel errors. Wrong dimensions/vector lengths, non-finite vectors, invalid IDs/budgets/options, closed indexes, and corrupt files return errors. Nil build contexts are treated as `context.Background()`. Constructors select the metric, avoiding a metric-less builder state. Construct public values through their constructors/open methods.

The public API intentionally excludes allocator/RNG/sorter/distance interfaces. Extend applications using small interfaces over the concrete operations they need. Internal metric specs, shared codecs/storage, and generated traversal keep implementation changes localized without hot-loop interface dispatch. Regenerate traversal with `go generate ./internal/engine`.

## Performance and development

Stable defaults use ARM64 NEON or AMD64 SSE2, with portable scalar fallback. Use `-tags goannoy_scalar` to force scalar kernels. Deterministic construction always uses scalar math. Experimental AMD64 Go SIMD is isolated behind **both** `GOEXPERIMENT=simd` and `-tags goannoy_experimental_simd`, pinned to Go 1.27.3. It is never needed for normal builds.

The reviewed workload dropped from 103 allocations per search to zero, with substantially lower search time and build memory. Exact commands, repeated measurements, platform details, and caveats are in [benchmark evidence](docs/release/v0.1.0-benchmarks.md).

```sh
make setup build
make test                 # weighted coverage: core/public >=90%, runtime >=85%
make test-race test-checkptr vet
go test -tags goannoy_scalar ./...
ANNOY_SOURCE=/path/to/pinned/annoy make compat
make bench
go run ./cmd/precision -items 10000 -length 128 -prec 100 -trees 10
```

Ordinary tests use committed independent fixtures and need no C++ compiler. The explicit compatibility gate requires the pinned source/compiler and fails if unavailable; see [compatibility harness](compat/README.md). CI executes stable platform, instrumentation, scalar, experimental, cross-build, and C++ gates. Timing history is non-blocking; allocation and coverage gates block.

Release scope and evidence: [v0.1.0 checklist](docs/release/v0.1.0-checklist.md), [design](docs/superpowers/specs/2026-11-09-goannoy-v0.1.0-design.md), and [implementation plan](docs/superpowers/plans/2026-11-09-goannoy-v0.1.0.md).
