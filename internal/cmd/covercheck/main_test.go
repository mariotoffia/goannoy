package main

import (
	"strings"
	"testing"
)

func TestCoverageWeightsStatementsAndMergesBlocks(t *testing.T) {
	profile := "mode: set\nexample/angular/a.go:1.1,2.1 9 1\nexample/angular/a.go:3.1,4.1 1 0\nexample/angular/a.go:1.1,2.1 9 1\nexample/cmd/a.go:1.1,2.1 10 0\n"
	core, overall, e := measure(strings.NewReader(profile))
	if e != nil {
		t.Fatal(e)
	}
	if core != 90 || overall != 45 {
		t.Fatalf("core=%g overall=%g", core, overall)
	}
}
