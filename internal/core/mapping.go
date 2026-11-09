package core

import (
	"fmt"
	goannoy "github.com/mariotoffia/goannoy"
	"os"
	"runtime"
	"unsafe"
)

type Mapping struct {
	Data    []byte
	release func([]byte) error
}

func OpenMapping(path string, options ...goannoy.OpenOption) (*Mapping, error) {
	var little uint32 = 1
	if *(*byte)(unsafe.Pointer(&little)) != 1 {
		return nil, goannoy.ErrUnsupportedPlatform
	}
	prefault := false
	for _, option := range options {
		if option != goannoy.Prefault {
			return nil, goannoy.ErrInvalidOption
		}
		prefault = true
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() < 1 || uint64(info.Size()) > uint64(^uint(0)>>1) {
		return nil, fmt.Errorf("%w: invalid file size", goannoy.ErrCorruptIndex)
	}
	m, err := mapFile(f, int(info.Size()))
	if err != nil {
		return nil, err
	}
	if prefault {
		var sum byte
		for n := 0; n < len(m.Data); n += os.Getpagesize() {
			sum ^= m.Data[n]
		}
		runtime.KeepAlive(sum)
	}
	return m, nil
}
func (m *Mapping) Close() error {
	if m.Data == nil {
		return nil
	}
	data := m.Data
	m.Data = nil
	if m.release != nil {
		return m.release(data)
	}
	return nil
}
