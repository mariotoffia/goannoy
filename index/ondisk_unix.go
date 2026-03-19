//go:build linux || darwin || freebsd || openbsd || netbsd

package index

import (
	"github.com/mariotoffia/goannoy/index/memory"
	"github.com/mariotoffia/goannoy/interfaces"
)

func newOnDiskBuildAllocator(fileName string) (interfaces.BuildIndexAllocator, error) {
	return memory.NewOnDiskBuildAllocator(fileName)
}
