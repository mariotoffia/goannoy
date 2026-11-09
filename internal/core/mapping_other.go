//go:build !(darwin || linux || freebsd || netbsd || openbsd || dragonfly)

package core

import (
	"io"
	"os"
)

func mapFile(f *os.File, size int) (*Mapping, error) {
	b := make([]byte, size)
	if _, err := io.ReadFull(f, b); err != nil {
		return nil, err
	}
	return &Mapping{Data: b}, nil
}
