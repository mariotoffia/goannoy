package compat

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const upstreamCommit = "379f744667aba6b40ba3db8a07678df173a88f74"

type fixture struct {
	Metric    string    `json:"metric"`
	Dimension int       `json:"dimension"`
	SHA256    string    `json:"sha256"`
	IDs       []int32   `json:"ids"`
	Distances []float32 `json:"distances"`
}
type manifest struct {
	Upstream string    `json:"upstream"`
	Seed     uint64    `json:"seed"`
	Trees    int       `json:"trees"`
	Items    int       `json:"items"`
	Fixtures []fixture `json:"fixtures"`
}

func readManifest(t *testing.T) manifest {
	t.Helper()
	b, err := os.ReadFile("fixtures/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var m manifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

// Catches missing or replaced independent oracles; fixture absence is a hard error.
func TestFixtureManifest(t *testing.T) {
	m := readManifest(t)
	if m.Upstream != upstreamCommit {
		t.Fatalf("upstream %q", m.Upstream)
	}
	if len(m.Fixtures) != 5 {
		t.Fatalf("got %d fixtures", len(m.Fixtures))
	}
	seen := map[string]bool{}
	for _, f := range m.Fixtures {
		t.Run(f.Metric, func(t *testing.T) {
			if seen[f.Metric] {
				t.Fatal("duplicate fixture")
			}
			seen[f.Metric] = true
			b, err := os.ReadFile(filepath.Join("fixtures", f.Metric+".ann"))
			if err != nil {
				t.Fatal(err)
			}
			got := sha256.Sum256(b)
			if hex.EncodeToString(got[:]) != f.SHA256 {
				t.Fatal("fixture digest changed")
			}
			if len(f.IDs) != 10 || len(f.Distances) != 10 {
				t.Fatal("query oracle missing")
			}
		})
	}
}
