---
name: goannoy-v01-upgrade
description: Use when upgrading an earlier GoAnnoy consumer to v0.1.0, especially generic builder.Index, direct index.New, GetNnsByVector, custom policy interfaces, or legacy Hamming files. Not for unrelated Annoy libraries or ordinary v0.1 usage.
---

# Upgrade GoAnnoy to v0.1.0

Target module: `github.com/mariotoffia/goannoy` (no `/v1`). This is an intentionally breaking pre-1.0 API. Historical tags are not reliably known: identify source shapes rather than inventing a supported version range.

## Inventory before editing

Find old imports, constructors, metric selection, vector/ID types, allocation policies, search contexts, batch concurrency, file provenance, and consumers of returned distances. Preserve unrelated changes and original index files.

Read [references/api-mapping.md](references/api-mapping.md) for source edits. For any Hamming consumer, also read [references/hamming.md](references/hamming.md). Inspect the target's public source when details differ; do not reconstruct deleted interfaces.

## Migration constraints

- Choose a concrete metric package: `angular`, `dot`, `euclidean`, `manhattan`, or `hamming`.
- Separate a mutable builder from its immutable result. Handle constructor, `Add`, build, open, search, and close errors.
- Check external IDs **before** converting to `goannoy.ItemID`. Builders accept `0 <= id < math.MaxInt32`; sparse IDs consume the addressable range.
- Float metrics accept `[]float32`. Float64 narrowing needs an explicit accuracy decision, finite/range checks, and application-level recall tests; it is not precision-preserving.
- Replace batch contexts with one reusable searcher per concurrent caller. Result buffers are caller-owned; copy results retained across buffer reuse. Share indexes, not searchers or builders.
- Preserve metric meaning: dot returns similarity (larger is better); the other metrics return distance.
- Standard files are headerless. `Open(path, dimension)` needs the correct metric, dimension, and compatible layout; successful validation cannot prove provenance.
- **Legacy GoAnnoy one-float-per-bit Hamming files require rebuilding from source vectors.** Do not claim a safe byte conversion. With missing sources, report a deployment blocker and retain the old reader/file.
- Custom allocator, sorter, RNG, node, and distance implementations have no drop-in public replacement. Record unresolved behavior instead of silently deleting requirements.

## Verify and report

Add focused regression tests before changing consumers: ID overflow, vector conversion, search ordering, scratch ownership, and file/rebuild behavior. Compile a representative consumer, then run its focused tests and full suite. In the target repository run:

```sh
make test
make test-race test-checkptr vet
ANNOY_SOURCE=/path/to/pinned/annoy make compat
```

Do not claim unavailable C++ gates passed. No publish, tag, merge, or destructive replacement is implied by this skill.

Deliver a migration report with automated edits, per-caller follow-ups, removed/custom behaviors, manual blockers, file provenance/rebuild requirements, and exact commands/results. Mark unverified deployment data explicitly.
