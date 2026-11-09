// Command covercheck enforces weighted runtime statement coverage.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	profile := flag.String("profile", "coverage.out", "Go coverage profile")
	coreGate := flag.Float64("core", 90, "minimum core/public coverage")
	overallGate := flag.Float64("overall", 85, "minimum overall runtime coverage")
	flag.Parse()
	f, e := os.Open(*profile)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	defer f.Close()
	core, overall, e := measure(f)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	fmt.Printf("core/public: %.2f%%; overall runtime: %.2f%%\n", core, overall)
	if core < *coreGate || overall < *overallGate {
		os.Exit(1)
	}
}
func measure(r io.Reader) (float64, float64, error) {
	type block struct {
		statements    int
		covered, core bool
	}
	blocks := map[string]block{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return 0, 0, fmt.Errorf("invalid profile line %q", line)
		}
		n, e := strconv.Atoi(fields[1])
		if e != nil {
			return 0, 0, e
		}
		count, e := strconv.Atoi(fields[2])
		if e != nil {
			return 0, 0, e
		}
		old := blocks[fields[0]]
		blocks[fields[0]] = block{n, old.covered || count > 0, !strings.Contains(fields[0], "/cmd/")}
	}
	if e := scanner.Err(); e != nil {
		return 0, 0, e
	}
	var total, covered, coreTotal, coreCovered int
	for _, b := range blocks {
		total += b.statements
		if b.covered {
			covered += b.statements
		}
		if b.core {
			coreTotal += b.statements
			if b.covered {
				coreCovered += b.statements
			}
		}
	}
	if total == 0 || coreTotal == 0 {
		return 0, 0, fmt.Errorf("empty runtime coverage profile")
	}
	return 100 * float64(coreCovered) / float64(coreTotal), 100 * float64(covered) / float64(total), nil
}
