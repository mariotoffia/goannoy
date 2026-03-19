package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mariotoffia/goannoy/builder"
	gorandom "github.com/mariotoffia/goannoy/random"
)

const upstreamTreeRelPath = "../.work/spotify-annoy/test/test.tree"

func TestMain(m *testing.M) {
	if err := os.MkdirAll("testdata", 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create testdata dir: %v\n", err)
		os.Exit(1)
	}

	copyUpstreamAngularFixture()
	generateDotFixture()

	os.Exit(m.Run())
}

// copyUpstreamAngularFixture copies the upstream C++ Annoy test.tree into
// testdata/ so binary compatibility tests can load it. If the source is
// not available (e.g. .work/ not checked out), the copy is silently
// skipped and dependent tests will skip themselves.
func copyUpstreamAngularFixture() {
	dst := filepath.Join("testdata", "test.tree")
	if _, err := os.Stat(dst); err == nil {
		return // already present
	}

	data, err := os.ReadFile(upstreamTreeRelPath)
	if err != nil {
		return // upstream .work/ not available; tests will skip
	}

	_ = os.WriteFile(dst, data, 0o644)
}

// generateDotFixture builds the dot-product fixture from the known dataset
// and deterministic seed so TestUpstreamDotBinaryCompatibility can load it.
func generateDotFixture() {
	dst := filepath.Join("testdata", "upstream_dot.tree")
	if _, err := os.Stat(dst); err == nil {
		return // already present
	}

	idx := builder.Index().
		Random(gorandom.NewKiss64Random(42)).
		DotProductDistance(3).
		SingleWorkerPolicy().
		Build()
	defer idx.Close()

	for i, v := range upstreamDotFixtureDataset {
		if err := idx.AddItem(int32(i), v); err != nil {
			fmt.Fprintf(os.Stderr, "dot fixture: AddItem %d: %v\n", i, err)
			return
		}
	}

	if err := idx.Build(3, 1); err != nil {
		fmt.Fprintf(os.Stderr, "dot fixture: Build: %v\n", err)
		return
	}

	if err := idx.Save(dst); err != nil {
		fmt.Fprintf(os.Stderr, "dot fixture: Save: %v\n", err)
	}
}
