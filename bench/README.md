# Reproducible benchmarks

Host: Apple M4 Pro, darwin/arm64. Toolchain: Go 1.27.3.
The pinned workload uses 10,000 x 128 normally distributed float32 vectors,
seed 99 for items, seed 1234 for 128 queries, build seed 42, ten trees,
one build worker, ten returned results, and automatic search_k.

Run:

```sh
go test ./bench -run '^$' -bench '^BenchmarkAngularSearch$/^dim128$' -benchmem -count=5 -benchtime=500ms
go test ./bench -run '^$' -bench '^BenchmarkAngularBuild$' -benchmem -count=5 -benchtime=300ms
go test ./bench -run TestSearchIntoAllocs
```

Scalar search runs: 22,259; 21,421; 20,586; 20,366; 20,473 ns/op.
Median: 20,586 ns/op, 0 B/op, 0 allocs/op.
Reviewed branch baseline: 36–46 microseconds, 2,542 B/op, 103 allocs/op.
The scalar median is 57.2% of the conservative 36-microsecond baseline.

Scalar build runs: 151.5; 156.1; 151.6; 273.9; 157.1 milliseconds.
Median: 156.1 milliseconds, 25.8 MB/op, about 6,241 allocs/op.
Reviewed baseline: 178–186 milliseconds, about 87.2 MB/op and 43,372 allocs/op.
One new build run was an outlier; this is why the evidence lists every run.

Compatibility is measured independently against pinned C++ query results.
Timing is a benchmark result, never a correctness assertion.

After direct metric-specific traversal and NEON integration, default accelerated
search runs were 8,110; 7,513; 7,533; 7,579; 9,700 ns/op.
Median: 7,579 ns/op, 0 B/op, 0 allocs/op. This is 34.5% of the reviewed
22-microsecond accelerated baseline and passes the 75% release gate.

Use `-tags goannoy_scalar` to force portable search kernels.
Stable defaults use ARM64 NEON or AMD64 SSE2; other platforms use scalar.
Deterministic builds always use scalar metric arithmetic.

Experimental AMD64 Go SIMD is pinned to Go 1.27.3 and requires both
`GOEXPERIMENT=simd` and `-tags goannoy_experimental_simd`.
Without the experiment, stable SSE2 remains available. Unsupported CPUs use
scalar in the experimental build. The normal build never imports archsimd.

Final verification (1s search samples) after scalar loop-control unrolling:
scalar 17,283;17,309;17,607;18,521;17,798ns, median17,607ns (48.9% of36µs).
NEON 9,334;8,210;8,541;8,367;9,107ns, median8,541ns (38.8% of22µs).
Both remain0B/op,0allocs/op. Final build median153.48ms,25.78MB/op,
about6,241allocs/op. See [release evidence](../docs/release/v0.1.0-benchmarks.md)
for all samples, footprint and caveats.
