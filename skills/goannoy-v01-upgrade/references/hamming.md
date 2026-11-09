# Hamming requires a representation decision

The old GoAnnoy Hamming implementation stored one float per bit. Those files are not Spotify-compatible and **must be rebuilt from the original bit semantics/vectors**. Missing source data is a real blocker. Do not reinterpret the bytes as packed words, overwrite the only file, or claim that successful `Open` proves compatibility. Retaining the old deployment is safer than inventing bit values.

The v0.1 `hamming` package takes a logical bit dimension and packed `[]uint64`:

```go
b, err := hamming.NewBuilder(130)
// Handle err; then Add(id, []uint64{word0, word1, word2}).
// Only bits0 and1 of word2 may be set.
```

Word count is `(bitDimension+63)/64`. Bit j is `words[j/64] & (1 << (j%64))`; unused high bits must be zero. This matches Spotify's Python-wrapper packing (values >0.5 become set bits), but apply that threshold to old data only after confirming that it expresses the application's intended semantics. Do not invent source vectors from the dimension alone.

For an independently verified standard Spotify Hamming file:

```go
idx, err := hamming.Open(path, 130)
if err != nil { return err }
defer idx.Close()
searcher := idx.NewSearcher()
n, err := searcher.SearchInto(queryWords, dst, goannoy.AutoSearchK)
```

The same provenance check applies even when file validation succeeds: Annoy has no metric/dimension header. Build a new filename, compare known queries with an independent popcount oracle and the old application, then separately authorize deployment replacement. Document source recovery or rebuilding as manual work when it cannot be performed.
