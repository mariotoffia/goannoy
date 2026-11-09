# Concurrency and ownership

An index is an immutable snapshot. Any number of goroutines may query it,
each with its own searcher. A searcher owns mutable scratch slices and must
have one caller at a time. Builders also have one caller at a time.

Each public index operation holds one read lock. Close acquires the write
lock, waits for admitted operations, releases storage, and rejects subsequent
queries with ErrClosed. Close is idempotent. There is no per-node locking.

Build accepts nil as context.Background. A canceled build returns the context
error and preserves both builder input and previous index snapshots.

Workers(1) uses Spotify's continuous single-threaded random stream and matches
the scalar fixture fields. Parallel builds seed each tree independently and
serialize trees in tree-number order. Scheduling does not change output.
Parallel files remain standard Annoy files; their topology can differ from a
one-worker build.

Balanced defaults to at most four workers, Throughput uses GOMAXPROCS, and
LowMemory uses one. Explicit Workers overrides the profile; Workers(0)
requests GOMAXPROCS. Profiles never change file layout or metric definitions.

OnDisk(path) uses a growing writable mapping while constructing trees, then
syncs and atomically commits the complete file. It uses one construction worker
to keep writable mapping ownership simple. On platforms without mmap, it uses
the portable memory-backed file path. Builder input vectors remain in memory.
