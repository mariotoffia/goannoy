// Command fixtures regenerates independent C++ fixtures and their manifest.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const commit = "379f744667aba6b40ba3db8a07678df173a88f74"

func main() {
	if err := generate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func generate() error {
	source := os.Getenv("ANNOY_SOURCE")
	if source == "" {
		return fmt.Errorf("set ANNOY_SOURCE to the pinned Spotify Annoy checkout")
	}
	out, err := exec.Command("git", "-C", source, "rev-parse", "HEAD").Output()
	if err != nil || strings.TrimSpace(string(out)) != commit {
		return fmt.Errorf("ANNOY_SOURCE must be at %s", commit)
	}
	dir, err := os.MkdirTemp("", "goannoy-compat-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	binary := filepath.Join(dir, "generate.out")
	cmd := exec.Command("c++", "-std=c++17", "-O2", "-ffp-contract=off", "-fno-tree-vectorize", "-I"+filepath.Join(source, "src"), "cpp/generate.cc", "-o", binary)
	if b, e := cmd.CombinedOutput(); e != nil {
		return fmt.Errorf("compile: %w: %s", e, b)
	}
	if err := os.MkdirAll("fixtures", 0755); err != nil {
		return err
	}
	b, err := exec.Command(binary, "write", "fixtures").Output()
	if err != nil {
		return err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	for _, v := range m["fixtures"].([]any) {
		f := v.(map[string]any)
		name := f["metric"].(string)
		data, e := os.ReadFile("fixtures/" + name + ".ann")
		if e != nil {
			return e
		}
		sum := sha256.Sum256(data)
		f["sha256"] = hex.EncodeToString(sum[:])
	}
	b, err = json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile("fixtures/manifest.json", append(b, '\n'), 0644); err != nil {
		return err
	}
	if out, err := exec.Command(binary, "read", "fixtures").Output(); err != nil {
		return err
	} else {
		fmt.Print(string(out))
	}
	return nil
}
