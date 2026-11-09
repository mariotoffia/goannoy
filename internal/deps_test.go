package internal_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeDependencyPolicy(t *testing.T) {
	cmd := exec.Command("go", "list", "-deps", "-f", "{{if not .Standard}}{{.ImportPath}}{{end}}", "./...")
	cmd.Dir = ".."
	out, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("%v: %s", e, out)
	}
	for _, name := range strings.Fields(string(out)) {
		if !strings.HasPrefix(name, "github.com/mariotoffia/goannoy") {
			t.Errorf("external dependency: %s", name)
		}
	}
}

func TestSingleMetricDoesNotLinkOtherEngines(t *testing.T) {
	for _, name := range []string{"angular", "dot", "euclidean", "manhattan", "hamming"} {
		t.Run(name, func(t *testing.T) { checkMetricFootprint(t, name) })
	}
}

func checkMetricFootprint(t *testing.T, name string) {
	t.Helper()
	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")
	binary := filepath.Join(dir, "consumer.out")
	vector := "[]float32{1,0}"
	if name == "hamming" {
		vector = "[]uint64{1}"
	}
	code := fmt.Sprintf(`package main
import("fmt";"github.com/mariotoffia/goannoy/%s")
func main(){b,_:=%s.NewBuilder(2);b.Add(0,%s);i,_:=b.Build(nil);defer i.Close();r,e:=i.NewSearcher().Search(%s,1,-1);fmt.Println(r,e)}
`, name, name, vector, vector)
	if e := os.WriteFile(source, []byte(code), 0600); e != nil {
		t.Fatal(e)
	}
	cmd := exec.Command("go", "build", "-o", binary, source)
	cmd.Dir = ".."
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("%v: %s", e, out)
	}
	out, e := exec.Command("go", "tool", "nm", binary).CombinedOutput()
	if e != nil {
		t.Fatalf("%v: %s", e, out)
	}
	for _, other := range []string{"Angular", "Dot", "Euclidean", "Manhattan", "Hamming"} {
		if strings.EqualFold(name, other) {
			continue
		}
		if strings.Contains(string(out), "(*Searcher).search"+other) {
			t.Errorf("%s consumer retains %s search engine", name, other)
		}
	}
	cmd = exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", binary, source)
	cmd.Dir = ".."
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("stripped build: %v: %s", e, out)
	}
	info, e := os.Stat(binary)
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("%s stripped consumer: %d bytes", name, info.Size())
}
