# Earlier source shapes → v0.1.0

| Earlier source | Replacement | Semantic check |
| --- | --- | --- |
| `builder.Index[float32,uint32]()` or `builder.Index()` | `<metric>.NewBuilder(dimension)` | Constructor returns builder and error |
| `AngularDistance`, `DotProductDistance`, `EuclideanDistance`, `ManhattanDistance` | `angular`, `dot`, `euclidean`, `manhattan` import | Metric is selected by package, not a runtime interface |
| Direct `index.New(...)` | Concrete package builder or `Open(path,dimension)` | Audit all old injected dependencies |
| `UseSingleWorkerPolicy()` | `Workers(1)` | Serial build matches pinned C++ meaningful bytes |
| `UseMultiWorkerPolicy()` | `Workers(0)` or explicit positive count | Zero uses GOMAXPROCS; parallel output is reproducible for the same worker count, not C++-byte-identical |
| Build `nTrees`, seed | `Trees(nTrees).Seed(seed)` before `Build(ctx)` | Nil context is accepted; cancellation returns an error |
| GC/mmap build allocator | `Build(ctx)` / `OnDisk(path)` | OnDisk tree construction is serial; input vectors remain in memory |
| Index mmap allocator / `Load` | `Open(path,dimension[,goannoy.Prefault])` | Immutable mapping; metric/layout provenance remains external |
| `Save` | `BuildTo(ctx,path)` | Builds a fresh complete file and atomically replaces where supported |
| `AddItem` | `Builder.Add(ItemID,vector)` | Input copied, duplicate IDs replace, holes are not valid items |
| `CreateContext()` | `Index.NewSearcher()` | Searcher owns reusable scratch and is not concurrency-safe |
| `GetNnsByVector(q,n,k,ctx)` | `SearchInto(q,dst,SearchK(k))` | `len(dst)` is requested count; consume only returned count |
| `GetNnsByItem` | `SearchByIDInto(id,dst,searchK)` | Holes/invalid IDs return errors |
| Separate IDs/distances | `[]goannoy.Result` | Dot is descending similarity; others ascending distance |
| Returned vector view | `Vector(id)` / `VectorInto(id,dst)` | Copies only; no mutable mmap views |
| `Unload`/`Close` | `Index.Close()` | Idempotent; coordinates with active operations |

Default configuration: 10 trees, Kiss64 seed `1234567890987654321`, Balanced profile. When Workers is omitted, Balanced caps workers at 4, Throughput uses GOMAXPROCS, LowMemory uses 1. Explicit Workers overrides profiles; `Workers(0)` uses GOMAXPROCS without the profile cap. Invalid fluent configuration is sticky: inspect `Err()` or handle the next `Add`/`Build` error. Invalid input data does not poison the builder.

```go
b, err := angular.NewBuilder(3)
if err != nil { return err }
b.Trees(10).Workers(0).Seed(42).Capacity(len(items))
for id, v := range items {
    if err := b.Add(goannoy.ItemID(id), v); err != nil { return err }
}
idx, err := b.Build(ctx)
if err != nil { return err }
defer idx.Close()
s := idx.NewSearcher()
dst := make([]goannoy.Result, 10)
n, err := s.SearchInto(query, dst, goannoy.AutoSearchK)
// Use dst[:n] only after checking err.
```

Imports and enclosing application function are omitted. Check IDs before casts in real adapters. `AutoSearchK` is -1, resolving to requested-neighbors × trees. Other budgets must be positive. `Search(q,limit,k)` is an allocating convenience wrapper.

`Items()` is maxID+1, not the number of inserted sparse items. Standard files support signed32 IDs and float32 vectors on little-endian systems. Public custom interfaces, float64 storage, and arbitrary ID widths were removed intentionally; do not recreate the old abstraction graph just to make compilation pass.
