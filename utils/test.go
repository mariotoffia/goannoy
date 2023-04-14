package utils

import (
	"os"
	"runtime"
	"strings"
)

var InTest bool

func init() {
	InTest = InTesting()
}

func InTesting() bool {
	suffix := ".test"
	if runtime.GOOS == "windows" {
		suffix += ".exe"
	}
	return len(os.Args) > 1 && strings.HasSuffix(os.Args[0], suffix) &&
		strings.HasPrefix(os.Args[1], "-test.")
}
