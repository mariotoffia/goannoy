package internal_test

import (
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
	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")
	binary := filepath.Join(dir, "consumer.out")
	code := `package main
import("fmt";"github.com/mariotoffia/goannoy/angular")
func main(){b,_:=angular.NewBuilder(2);b.Add(0,[]float32{1,0});i,_:=b.Build(nil);defer i.Close();r,e:=i.NewSearcher().Search([]float32{1,0},1,-1);fmt.Println(r,e)}
`
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
	for _, symbol := range []string{"(*Searcher).searchHamming", "(*Searcher).searchDot", "(*Searcher).searchManhattan", "(*Searcher).searchEuclidean", "metric.splitHamming"} {
		if strings.Contains(string(out), symbol) {
			t.Errorf("angular consumer retains %s", symbol)
		}
	}
}
