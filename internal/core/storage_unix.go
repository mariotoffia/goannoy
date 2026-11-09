//go:build darwin || linux || freebsd || netbsd || openbsd || dragonfly

package core

import (
	"os"
	"syscall"
)

func mapWritable(f *os.File, size int) ([]byte, error) {
	return syscall.Mmap(int(f.Fd()), 0, size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
}
func unmapWritable(_ *os.File, data []byte, _ int) error { return syscall.Munmap(data) }
func syncDirectory(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
