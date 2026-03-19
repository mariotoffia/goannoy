package vector

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
	"unsafe"
)

// TestDotFloat32ScalarVsAccelerated runs a high-volume dot-product workload
// through both the scalar and the active accelerated backend, reporting
// wall-clock elapsed time for each.
//
// Without -tags accelerate, only the scalar baseline is measured.
// With -tags accelerate, both are measured and compared side by side.
//
// Run:
//
//	go test -tags accelerate -v -run TestDotFloat32ScalarVsAccelerated ./vector/
func TestDotFloat32ScalarVsAccelerated(t *testing.T) {
	backend := activeDotFloat32BackendName()
	accelerated := backend != "scalar"

	type scenario struct {
		dim   int
		pairs int
	}

	scenarios := []scenario{
		{dim: 128, pairs: 2_000_000},
		{dim: 512, pairs: 1_000_000},
		{dim: 1536, pairs: 500_000},
	}

	type result struct {
		dim     int
		pairs   int
		scalar  time.Duration
		accel   time.Duration
		speedup float64
	}

	results := make([]result, 0, len(scenarios))

	for _, sc := range scenarios {
		t.Run(fmt.Sprintf("dim=%d", sc.dim), func(t *testing.T) {
			pool := makeVectorPool(sc.dim)

			scalarElapsed := runDotBatch(pool, sc.dim, sc.pairs, dotFloat32UnsafeScalar)

			var accelElapsed time.Duration
			if accelerated {
				accelElapsed = runDotBatch(pool, sc.dim, sc.pairs, dotFloat32BackendUnsafeFn)
			}

			r := result{
				dim:    sc.dim,
				pairs:  sc.pairs,
				scalar: scalarElapsed,
				accel:  accelElapsed,
			}

			if accelerated {
				r.speedup = float64(scalarElapsed) / float64(accelElapsed)
				t.Logf("dim=%-5d  scalar=%-10s  %s=%-10s  speedup=%.2fx",
					sc.dim, scalarElapsed.Truncate(time.Millisecond),
					backend, accelElapsed.Truncate(time.Millisecond), r.speedup)
			} else {
				t.Logf("dim=%-5d  scalar=%-10s  (no accelerated backend)", sc.dim, scalarElapsed.Truncate(time.Millisecond))
			}

			results = append(results, r)
		})
	}

	// Summary table.
	t.Log("")
	t.Logf("====== Summary (backend: %s) ======", backend)
	t.Log("")

	if accelerated {
		hdr := fmt.Sprintf("  %-8s  %-12s  %-12s  %s", "Dim", "Scalar", backend, "Speedup")
		t.Log(hdr)
		t.Log("  " + strings.Repeat("-", len(hdr)-2))
		for _, r := range results {
			t.Logf("  %-8d  %-12s  %-12s  %.2fx",
				r.dim,
				r.scalar.Truncate(time.Millisecond),
				r.accel.Truncate(time.Millisecond),
				r.speedup,
			)
		}
	} else {
		hdr := fmt.Sprintf("  %-8s  %s", "Dim", "Scalar")
		t.Log(hdr)
		t.Log("  " + strings.Repeat("-", len(hdr)-2))
		for _, r := range results {
			t.Logf("  %-8d  %s", r.dim, r.scalar.Truncate(time.Millisecond))
		}
		t.Log("")
		t.Log("  No accelerated backend active. To compare against NEON, run:")
		t.Log("    go test -tags accelerate -v -run TestDotFloat32ScalarVsAccelerated ./vector/")
	}
}

func makeVectorPool(dim int) [][]float32 {
	const poolSize = 1024
	rng := rand.New(rand.NewSource(int64(dim)))
	pool := make([][]float32, poolSize)

	for i := range pool {
		v := make([]float32, dim)
		for j := range v {
			v[j] = rng.Float32()*2 - 1
		}
		pool[i] = v
	}

	return pool
}

func runDotBatch(
	pool [][]float32,
	dim, pairs int,
	fn func(a, b *float32, n int) float32,
) time.Duration {
	poolSize := len(pool)
	var sink float32

	start := time.Now()
	for i := range pairs {
		a := pool[i%poolSize]
		b := pool[(i+1)%poolSize]
		sink += fn(
			(*float32)(unsafe.Pointer(unsafe.SliceData(a))),
			(*float32)(unsafe.Pointer(unsafe.SliceData(b))),
			dim,
		)
	}
	elapsed := time.Since(start)

	// Use sink to prevent dead-code elimination.
	if sink == 31415926 {
		fmt.Println(sink)
	}

	return elapsed
}
