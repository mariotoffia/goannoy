//go:build cppcompat

package compat

import (
	"encoding/json"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/engine"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// This explicit release gate requires C++ and the pinned checkout, never skips.
func TestCppCompatibilityReadsGoBuilds(t *testing.T) {
	source := os.Getenv("ANNOY_SOURCE")
	if source == "" {
		t.Fatal("cppcompat requires ANNOY_SOURCE pointing at the pinned checkout")
	}
	version, e := exec.Command("git", "-C", source, "rev-parse", "HEAD").Output()
	if e != nil || strings.TrimSpace(string(version)) != upstreamCommit {
		t.Fatal("ANNOY_SOURCE is not the pinned commit")
	}
	executable := filepath.Join(t.TempDir(), "reader.out")
	cmd := exec.Command("c++", "-std=c++17", "-O2", "-ffp-contract=off", "-fno-tree-vectorize", "-I"+filepath.Join(source, "src"), "cpp/generate.cc", "-o", executable)
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("compile %v: %s", e, out)
	}
	want := readManifest(t)
	for _, workers := range []int{1, 2} {
		dir := t.TempDir()
		for _, k := range []codec.Kind{codec.Angular, codec.Dot, codec.Euclidean, codec.Manhattan, codec.Hamming} {
			dim := 3
			if k == codec.Hamming {
				dim = 130
			}
			b, e := engine.NewBuilder(k, dim)
			if e != nil {
				t.Fatal(e)
			}
			b.Config.Trees(3).Workers(workers).Seed(42)
			for id := 0; id < 64; id++ {
				if k == codec.Hamming {
					n := uint64(id)
					e = b.AddWords(goannoy.ItemID(id), []uint64{n * 0x9e3779b97f4a7c15, (n + 3) * 0xd6e8feb86659fd93, n & 3})
				} else {
					e = b.AddFloat(goannoy.ItemID(id), []float32{float32((id*17)%97-48) / 8, float32((id*31)%89-44) / 8, float32((id*43)%83-41) / 8})
				}
				if e != nil {
					t.Fatal(e)
				}
			}
			i, e := b.BuildTo(nil, filepath.Join(dir, k.String()+".ann"))
			if e != nil {
				t.Fatal(e)
			}
			if e := i.Close(); e != nil {
				t.Fatal(e)
			}
		}
		out, e := exec.Command(executable, "read", dir).CombinedOutput()
		if e != nil {
			t.Fatalf("C++ read %v: %s", e, out)
		}
		var got manifest
		if e := json.Unmarshal(out, &got); e != nil {
			t.Fatalf("%v: %s", e, out)
		}
		for n, f := range got.Fixtures {
			for j, id := range f.IDs {
				if id != want.Fixtures[n].IDs[j] || math.Abs(float64(f.Distances[j]-want.Fixtures[n].Distances[j])) > 1e-5 {
					t.Fatalf("workers=%d metric=%s rank=%d", workers, f.Metric, j)
				}
			}
		}
	}
}
