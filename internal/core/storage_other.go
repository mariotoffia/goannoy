//go:build !(darwin || linux || freebsd || netbsd || openbsd || dragonfly)

package core

import (
	"io"
	"os"
)

func mapWritable(f *os.File, size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := f.ReadAt(b, 0)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return b, nil
}
func unmapWritable(f *os.File, data []byte, used int) error {
	_, err := f.WriteAt(data[:used], 0)
	return err
}
func syncDirectory(string) error { return nil }
