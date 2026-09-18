# GoAnnoy Branch Audit and Repair Design

## Context

The `buf/bug-fixes` branch is four commits ahead of `main` and changes 68 files. Despite its name, it combines three substantial areas:

1. Annoy algorithm and binary-compatibility corrections.
2. Optional ARM64 and AMD64 dot-product acceleration.
3. On-disk index construction and file-backed memory lifecycle changes.

There is no pull request or tracking issue for the branch. The baseline portable test command, `go test ./...`, passes on Darwin/ARM64 with Go 1.27.1. The baseline `go vet ./...` command fails, including an ARM64 assembly declaration diagnostic and unkeyed `interfaces.Pair` literals. The repository has no GitHub Actions workflow, so the branch currently has no remote verification matrix.

## Goal

Establish whether every existing branch fix is behaviorally correct, find additional defects introduced or exposed by the branch, repair confirmed defects with strict test-driven development, and publish an evidence-backed GitHub issue and pull request whose issue table accurately records the result.

## Non-goals

- Do not redesign the public index API unless a confirmed correctness defect cannot be fixed compatibly.
- Do not perform unrelated formatting or broad refactoring.
- Do not claim runtime verification for a platform that was only cross-compiled.
- Do not add benchmark-driven optimizations unless the audit identifies a correctness regression caused by the current optimization.
- Do not treat an existing passing test as proof of a newly discovered defect; each new repair requires a test observed failing against the pre-fix implementation.

## Audit Areas

### Annoy compatibility and search correctness

Review the changes to distance implementations, node layout, split construction, item insertion, tree building, candidate selection, distance conversion, serialization, and loading. Compare externally visible behavior with the repository's interfaces and with the corresponding Spotify Annoy implementation where compatibility is claimed.

Key cases include sparse item identifiers, negative identifiers, zero or small indexes, duplicate vectors, seed determinism, multithreaded construction, context reuse, search buffer limits, `search_k`, saved-index round trips, and upstream binary fixtures.

### CPU acceleration

Review build constraints, backend selection, scalar fallback, unsafe slice handling, vector tails, alignment assumptions, assembly declarations, numerical tolerance, empty and unequal inputs, and architecture-specific build requirements.

Runtime tests will cover the local ARM64 scalar and accelerated paths. AMD64 accelerated code will be cross-compiled locally; it must not be described as runtime-tested unless an AMD64 runner is actually available.

### On-disk construction and memory lifecycle

Review allocator growth, truncation, mmap remapping, synchronization, prefaulting, file ownership, temporary-file behavior, failure cleanup, `BuildOnDisk`, `Save`, `Load`, `Unload`, and repeat-operation behavior.

Tests must assert consumer-visible behavior and filesystem effects using temporary directories. They must not inspect private source text or add production-only cleanup methods for tests.

### Repository and delivery quality

Review public documentation, build commands, module/toolchain requirements, committed local configuration, generated artifacts, vet diagnostics, and the absence of automated checks insofar as these affect the correctness or reproducibility of this branch.

## Investigation Method

For each changed subsystem:

1. Read the complete branch diff and the relevant base implementation.
2. Reconstruct the intended behavior from public interfaces, current tests, commit messages, and upstream Annoy semantics.
3. Trace suspicious values and resource ownership from public entry point to storage or result.
4. Record one concrete hypothesis at a time.
5. Reproduce the suspected fault consistently before proposing a fix.
6. Classify the result as confirmed, already correctly fixed, false positive, or unresolved due to unavailable platform/runtime evidence.

Severity uses the following definitions:

- **Critical:** corrupts persistent data or creates a broadly exploitable unsafe-memory failure.
- **High:** returns incorrect public results, crashes valid usage, leaks or corrupts files, or breaks a supported platform.
- **Medium:** violates a documented edge-case contract, produces nondeterministic behavior, or makes supported verification/build workflows fail.
- **Low:** delivery, documentation, or maintainability defect with limited runtime impact.

## TDD Repair Protocol

Every confirmed production defect follows an independent red-green-refactor cycle:

1. Name the exact production mutation or wrong branch the test detects.
2. Add the smallest test that exercises the real public or package boundary with hand-derived expected values.
3. Run the focused test and confirm it fails for the predicted behavioral reason rather than a setup or compilation error.
4. Implement the smallest root-cause correction.
5. Re-run the focused test and its package tests.
6. Refactor only after the tests are green.
7. Run the broader verification matrix before marking the issue fixed.

When an existing branch change appears to be a fix, temporarily comparing or reverting the relevant production hunk is acceptable as an audit technique, but the branch history will not be destructively rewritten. New regression tests must still demonstrate that the faulty behavior is detectable.

## Verification Matrix

The final verification will use the commands that are applicable to the host and code selected by build tags:

```bash
go test ./...
go test -race ./...
go test -tags accelerate ./...
go vet ./...
go build ./...
GOOS=linux GOARCH=amd64 GOEXPERIMENT=arenas,simd go build -tags accelerate ./...
GOOS=linux GOARCH=arm64 GOEXPERIMENT=arenas go build -tags accelerate ./...
git diff --check origin/main...HEAD
```

Focused tests will be run first during each TDD cycle. The final report will distinguish runtime tests, race/vet analysis, and cross-compilation. A command that cannot run will be reported with its exact blocker instead of being silently omitted.

## GitHub Workflow

After the implementation plan is approved:

1. Create one GitHub issue describing the audit scope, success criteria, and verification matrix.
2. Create a draft pull request from `buf/bug-fixes` to `main` containing `Closes #<issue-number>`.
3. Add an issue ledger to the pull request body with these columns:

   | ID | Area | Severity | Short description | Fixed? | Regression test / evidence |
   |---|---|---|---|---|---|

4. Add a row for every confirmed issue and every material existing fix whose correctness was explicitly evaluated. Use `Yes`, `No`, or `N/A — existing fix validated` in the **Fixed?** column.
5. Keep descriptions short, but link each status to a test name, verification command, commit, or explicit limitation.
6. Push each coherent TDD repair as a normal commit without rewriting the four existing commits.
7. Update the issue and PR table after final verification. Mark the PR ready only if no unresolved correctness defect remains; otherwise leave it draft and keep the tracking issue open by changing `Closes` to `Refs`.

All GitHub reads and writes will use the pre-authenticated `gh` CLI.

## Deliverables

- A committed regression test for every newly repaired production defect.
- Minimal production fixes that pass their focused tests and the applicable verification matrix.
- A GitHub issue tracking the audit.
- A linked pull request containing the complete issue-status table.
- A final review summary that separates validated existing fixes, newly fixed defects, unresolved limitations, and non-issues.

## Success Criteria

The work is complete only when:

1. Every changed subsystem has been reviewed against its intended contract.
2. Every confirmed defect is represented in the PR ledger.
3. Every newly fixed production defect has a recorded failing-then-passing regression test.
4. Every `Fixed?` value is supported by fresh verification evidence.
5. The final branch diff contains no unexplained local settings or generated artifacts.
6. The GitHub issue and PR link to each other and accurately reflect whether unresolved work remains.
