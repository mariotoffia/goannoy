package tests

// Tests ported from the original Spotify Annoy test suite:
// https://github.com/spotify/annoy/tree/main/test
//
// Only angular-distance tests are ported since that is the primary
// distance metric supported by the Go implementation.

import (
	"math"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Ported from angular_index_test.py ---

// test_get_nns_by_vector
func TestUpstreamAngularGetNnsByVector(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	idx.AddItem(0, []float32{0, 0, 1})
	idx.AddItem(1, []float32{0, 1, 0})
	idx.AddItem(2, []float32{1, 0, 0})
	idx.Build(10, -1)

	ctx := idx.CreateContext()

	r, _ := idx.GetNnsByVector([]float32{3, 2, 1}, 3, -1, ctx)
	assert.Equal(t, []uint32{2, 1, 0}, r)

	r, _ = idx.GetNnsByVector([]float32{1, 2, 3}, 3, -1, ctx)
	assert.Equal(t, []uint32{0, 1, 2}, r)

	r, _ = idx.GetNnsByVector([]float32{2, 0, 1}, 3, -1, ctx)
	assert.Equal(t, []uint32{2, 0, 1}, r)
}

// test_get_nns_by_item
func TestUpstreamAngularGetNnsByItem(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	idx.AddItem(0, []float32{2, 1, 0})
	idx.AddItem(1, []float32{1, 2, 0})
	idx.AddItem(2, []float32{0, 0, 1})
	idx.Build(10, -1)

	ctx := idx.CreateContext()

	r, _ := idx.GetNnsByItem(0, 3, -1, ctx)
	assert.Equal(t, []uint32{0, 1, 2}, r)

	r, _ = idx.GetNnsByItem(1, 3, -1, ctx)
	assert.Equal(t, []uint32{1, 0, 2}, r)
}

// test_dist: angular distance between [0,1] and [1,1]
func TestUpstreamAngularDist(t *testing.T) {
	idx := angularIdx(2)
	defer idx.Close()

	idx.AddItem(0, []float32{0, 1})
	idx.AddItem(1, []float32{1, 1})
	idx.Build(10, -1)

	d := idx.GetDistance(0, 1)
	expected := float32(math.Sqrt(2.0 * (1.0 - 1.0/math.Sqrt(2.0))))
	assert.InDelta(t, expected, d, 1e-5)
}

// test_dist_2: parallel vectors should have distance ~0
func TestUpstreamAngularDistParallel(t *testing.T) {
	idx := angularIdx(2)
	defer idx.Close()

	idx.AddItem(0, []float32{1000, 0})
	idx.AddItem(1, []float32{10, 0})
	idx.Build(10, -1)

	d := idx.GetDistance(0, 1)
	assert.InDelta(t, 0.0, d, 1e-5)
}

// test_dist_3: angular distance formula verification
func TestUpstreamAngularDistFormula(t *testing.T) {
	idx := angularIdx(2)
	defer idx.Close()

	idx.AddItem(0, []float32{97, 0})
	idx.AddItem(1, []float32{42, 42})
	idx.Build(10, -1)

	d := idx.GetDistance(0, 1)
	invSqrt2 := 1.0 / math.Sqrt(2.0)
	expected := math.Sqrt(math.Pow(1-invSqrt2, 2) + math.Pow(invSqrt2, 2))
	assert.InDelta(t, expected, float64(d), 1e-3)
}

// test_dist_degen: distance to zero vector
func TestUpstreamAngularDistDegen(t *testing.T) {
	idx := angularIdx(2)
	defer idx.Close()

	idx.AddItem(0, []float32{1, 0})
	idx.AddItem(1, []float32{0, 0})
	idx.Build(10, -1)

	d := idx.GetDistance(0, 1)
	assert.InDelta(t, math.Sqrt(2.0), float64(d), 1e-3)
}

// test_get_nns_search_k: explicit search_k parameter
func TestUpstreamAngularGetNnsSearchK(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	idx.AddItem(0, []float32{0, 0, 1})
	idx.AddItem(1, []float32{0, 1, 0})
	idx.AddItem(2, []float32{1, 0, 0})
	idx.Build(10, -1)

	ctx := idx.CreateContext()

	r, _ := idx.GetNnsByItem(0, 3, 10, ctx)
	assert.Equal(t, []uint32{0, 1, 2}, r)

	r, _ = idx.GetNnsByVector([]float32{3, 2, 1}, 3, 10, ctx)
	assert.Equal(t, []uint32{2, 1, 0}, r)
}

// test_include_dists: opposite vectors have distance ~2.0
func TestUpstreamAngularIncludeDists(t *testing.T) {
	f := 40
	idx := angularIdx(f)
	defer idx.Close()

	v := make([]float32, f)
	negV := make([]float32, f)
	rng := rand.New(rand.NewSource(42))
	for i := range v {
		v[i] = float32(rng.NormFloat64())
		negV[i] = -v[i]
	}

	idx.AddItem(0, v)
	idx.AddItem(1, negV)
	idx.Build(10, -1)

	ctx := idx.CreateContext()
	indices, dists := idx.GetNnsByItem(0, 2, 10, ctx)
	assert.Equal(t, []uint32{0, 1}, indices)
	assert.InDelta(t, 0.0, dists[0], 1e-3)
	assert.InDelta(t, 2.0, dists[1], 1e-3)
}

// test_only_one_item: single item, high dim, save/load roundtrip
func TestUpstreamAngularOnlyOneItem(t *testing.T) {
	f := 100
	dir := t.TempDir()
	path := filepath.Join(dir, "foo.ann")

	idx := angularIdxMmap(f)
	defer idx.Close()

	v := make([]float32, f)
	rng := rand.New(rand.NewSource(42))
	for i := range v {
		v[i] = float32(rng.NormFloat64())
	}

	idx.AddItem(0, v)
	idx.Build(10, -1)
	err := idx.Save(path)
	require.NoError(t, err)

	idx2 := angularIdxMmap(f)
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	ctx := idx2.CreateContext()
	query := make([]float32, f)
	for i := range query {
		query[i] = float32(rng.NormFloat64())
	}

	result, _ := idx2.GetNnsByVector(query, 50, -1, ctx)
	assert.Equal(t, []uint32{0}, result)
}

// test_single_vector: single vector with distances
func TestUpstreamAngularSingleVector(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "1.ann")

	idx := angularIdxMmap(3)
	defer idx.Close()

	idx.AddItem(0, []float32{1, 0, 0})
	idx.Build(10, -1)
	err := idx.Save(path)
	require.NoError(t, err)

	ctx := idx.CreateContext()
	indices, dists := idx.GetNnsByVector([]float32{1, 0, 0}, 3, -1, ctx)
	assert.Equal(t, []uint32{0}, indices)
	assert.InDelta(t, 0.0, dists[0]*dists[0], 1e-5)
}

// test_load_save_get_item_vector: vectors preserved through save/load
func TestUpstreamAngularLoadSaveGetItemVector(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "blah.ann")

	idx := angularIdxMmap(3)
	defer idx.Close()

	idx.AddItem(0, []float32{1.1, 2.2, 3.3})
	idx.AddItem(1, []float32{4.4, 5.5, 6.6})
	idx.AddItem(2, []float32{7.7, 8.8, 9.9})

	v0 := idx.GetItem(0)
	assertVecAlmostEqual(t, []float32{1.1, 2.2, 3.3}, v0)

	idx.Build(10, -1)
	err := idx.Save(path)
	require.NoError(t, err)

	v1 := idx.GetItem(1)
	assertVecAlmostEqual(t, []float32{4.4, 5.5, 6.6}, v1)

	idx2 := angularIdxMmap(3)
	defer idx2.Close()

	err = idx2.Load(path)
	require.NoError(t, err)

	v2 := idx2.GetItem(2)
	assertVecAlmostEqual(t, []float32{7.7, 8.8, 9.9}, v2)
}

// test_item_vector_after_save: vectors accessible after save (#279)
func TestUpstreamAngularItemVectorAfterSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "something.ann")

	idx := angularIdxMmap(3)
	defer idx.Close()

	idx.AddItem(0, []float32{0, 0, 0}) // placeholder
	idx.AddItem(1, []float32{1, 0, 0})
	idx.AddItem(2, []float32{0, 1, 0})
	idx.AddItem(3, []float32{0, 0, 1})
	idx.Build(10, -1)

	v3 := idx.GetItem(3)
	assertVecAlmostEqual(t, []float32{0, 0, 1}, v3)

	ctx := idx.CreateContext()
	r, _ := idx.GetNnsByItem(1, 999, -1, ctx)
	assertContainsAll(t, r, []uint32{1, 2, 3})

	err := idx.Save(path)
	require.NoError(t, err)

	v3After := idx.GetItem(3)
	assertVecAlmostEqual(t, []float32{0, 0, 1}, v3After)

	ctx2 := idx.CreateContext()
	r2, _ := idx.GetNnsByItem(1, 999, -1, ctx2)
	assertContainsAll(t, r2, []uint32{1, 2, 3})
}

// test_get_lots_of_nns: requesting more NNs than items returns all items
func TestUpstreamGetLotsOfNns(t *testing.T) {
	f := 10
	idx := angularIdx(f)
	defer idx.Close()

	rng := rand.New(rand.NewSource(42))
	v := make([]float32, f)
	for i := range v {
		v[i] = float32(rng.NormFloat64())
	}
	idx.AddItem(0, v)
	idx.Build(10, -1)

	ctx := idx.CreateContext()
	for j := 0; j < 100; j++ {
		r, _ := idx.GetNnsByItem(0, 999999, -1, ctx)
		assert.Equal(t, []uint32{0}, r)
	}
}

// test_save_without_build: saving without calling build must error
func TestUpstreamSaveWithoutBuild(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.ann")

	idx := angularIdxMmap(10)
	defer idx.Close()

	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 100; i++ {
		v := make([]float32, 10)
		for j := range v {
			v[j] = float32(rng.NormFloat64())
		}
		idx.AddItem(uint32(i), v)
	}

	err := idx.Save(path)
	assert.Error(t, err, "saving unbuilt index should error")
}

// test_build_twice: building twice must panic
func TestUpstreamBuildTwice(t *testing.T) {
	idx := angularIdx(10)
	defer idx.Close()

	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 100; i++ {
		v := make([]float32, 10)
		for j := range v {
			v[j] = float32(rng.NormFloat64())
		}
		idx.AddItem(uint32(i), v)
	}
	idx.Build(10, -1)

	assert.Panics(t, func() {
		idx.Build(10, -1)
	}, "building twice should panic")
}

// test_large_index: 10k items, paired close points. Approximate search so
// we check precision (>= 90%) rather than requiring exact matches.
func TestUpstreamAngularLargeIndex(t *testing.T) {
	f := 10
	nItems := 10000
	idx := angularIdx(f)
	defer idx.Close()

	rng := rand.New(rand.NewSource(42))

	for j := 0; j < nItems; j += 2 {
		p := make([]float32, f)
		for z := range p {
			p[z] = float32(rng.NormFloat64())
		}

		f1 := float32(rng.Float64()) + 1
		f2 := float32(rng.Float64()) + 1

		x := make([]float32, f)
		y := make([]float32, f)
		for z := range p {
			x[z] = f1*p[z] + float32(rng.NormFloat64())*1e-2
			y[z] = f2*p[z] + float32(rng.NormFloat64())*1e-2
		}

		idx.AddItem(uint32(j), x)
		idx.AddItem(uint32(j+1), y)
	}

	idx.Build(10, -1)

	ctx := idx.CreateContext()
	correct := 0
	total := 0
	searchK := nItems // generous search budget

	for j := 0; j < nItems; j += 2 {
		r, _ := idx.GetNnsByItem(uint32(j), 2, searchK, ctx)
		if len(r) >= 2 && r[0] == uint32(j) && r[1] == uint32(j+1) {
			correct++
		}
		total++

		r2, _ := idx.GetNnsByItem(uint32(j+1), 2, searchK, ctx)
		if len(r2) >= 2 && r2[0] == uint32(j+1) && r2[1] == uint32(j) {
			correct++
		}
		total++
	}

	precision := float64(correct) / float64(total)
	assert.GreaterOrEqual(t, precision, 0.90,
		"paired close-point precision should be >= 90%%, got %.2f%%", precision*100)
}

// test_distance_consistency: verify GetDistance matches NNS distances and
// the manual angular distance formula.
func TestUpstreamAngularDistanceConsistency(t *testing.T) {
	n, f := 1000, 3
	idx := angularIdx(f)
	defer idx.Close()

	rng := rand.New(rand.NewSource(42))

	for j := 0; j < n; j++ {
		var v []float32
		for {
			v = make([]float32, f)
			dot := float32(0)
			for z := range v {
				v[z] = float32(rng.NormFloat64())
				dot += v[z] * v[z]
			}
			if dot > 0.1 {
				break
			}
		}
		idx.AddItem(uint32(j), v)
	}

	idx.Build(10, -1)

	ctx := idx.CreateContext()

	for trial := 0; trial < 100; trial++ {
		a := uint32(rng.Intn(n))
		indices, dists := idx.GetNnsByItem(a, 100, -1, ctx)

		for k, b := range indices {
			dist := dists[k]
			d2 := idx.GetDistance(a, b)
			assert.InDelta(t, float64(d2), float64(dist), 1e-3,
				"GetDistance(%d,%d) should match NNS distance", a, b)

			u := idx.GetItem(a)
			v := idx.GetItem(b)
			manual := manualAngularDist(u, v)
			assert.InDelta(t, float64(manual), float64(dist), 1e-3,
				"manual distance for (%d,%d) should match", a, b)
		}
	}
}

// --- Save/Load completeness tests ---

// test_save_twice: saving to two different files
func TestUpstreamSaveTwice(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, "t1.ann")
	path2 := filepath.Join(dir, "t2.ann")

	idx := angularIdxMmap(10)
	defer idx.Close()

	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 100; i++ {
		v := make([]float32, 10)
		for j := range v {
			v[j] = float32(rng.NormFloat64())
		}
		idx.AddItem(uint32(i), v)
	}
	idx.Build(10, -1)

	err := idx.Save(path1)
	require.NoError(t, err)

	// Save+Load to path1, now save again to path2
	err = idx.Save(path2)
	require.NoError(t, err)
}

// test_get_nns_with_distances: verify NNS distances are correct for angular
func TestUpstreamAngularGetNnsWithDistances(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	idx.AddItem(0, []float32{0, 0, 1})
	idx.AddItem(1, []float32{0, 1, 0})
	idx.AddItem(2, []float32{1, 0, 0})
	idx.Build(10, -1)

	ctx := idx.CreateContext()
	l, d := idx.GetNnsByItem(0, 3, -1, ctx)

	assert.Equal(t, []uint32{0, 1, 2}, l)
	assert.InDelta(t, 0.0, d[0], 1e-5, "distance to self should be ~0")
	// Orthogonal vectors have angular distance = sqrt(2)
	assert.InDelta(t, math.Sqrt(2.0), float64(d[1]), 1e-3)
	assert.InDelta(t, math.Sqrt(2.0), float64(d[2]), 1e-3)
}

