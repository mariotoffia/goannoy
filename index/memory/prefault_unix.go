//go:build linux || darwin || freebsd || openbsd || netbsd

package memory

import (
	"syscall"
	"unsafe"
)

func madviseWillNeed(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	_, _, e1 := syscall.Syscall(
		syscall.SYS_MADVISE,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		uintptr(syscall.MADV_WILLNEED),
	)
	if e1 != 0 {
		return e1
	}
	return nil
}

func (mi *mmapIndexAllocation) Prefault() error {
	if mi.data == nil {
		return nil
	}
	return madviseWillNeed(mi.data)
}
