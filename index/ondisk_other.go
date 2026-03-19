//go:build !(linux || darwin || freebsd || openbsd || netbsd)

package index

import (
	"errors"

	"github.com/mariotoffia/goannoy/interfaces"
)

func newOnDiskBuildAllocator(_ string) (interfaces.BuildIndexAllocator, error) {
	return nil, errors.New("on-disk build is not supported on this platform")
}
