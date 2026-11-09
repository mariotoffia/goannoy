//go:build darwin || linux || freebsd || netbsd || openbsd || dragonfly

package core

import (
	"os"
	"syscall"
)

func mapFile(f *os.File, size int) (*Mapping, error) {
	b, err := syscall.Mmap(int(f.Fd()), 0, size, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	return &Mapping{Data: b, release: syscall.Munmap}, nil
}
