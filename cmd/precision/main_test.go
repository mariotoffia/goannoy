package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrecisionUsesIndependentExactOracle(t *testing.T) {
	var out bytes.Buffer
	if e := run([]string{"-items", "16", "-length", "3", "-prec", "4", "-trees", "3", "-search-k", "10000"}, &out); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(out.String(), "recall@10=1.0000") {
		t.Fatal(out.String())
	}
	if e := run([]string{"-items", "0"}, &out); e == nil {
		t.Fatal("accepted empty workload")
	}
}
