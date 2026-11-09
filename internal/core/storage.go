package core

import (
	"os"
	"path/filepath"
)

// AtomicWrite commits a complete sibling file, preserving existing mappings.
// The callback must leave the file open. On callback/sync/rename failure the
// original destination survives; a directory-sync error occurs after commit.
func AtomicWrite(path string, write func(*os.File) error) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".goannoy-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	defer f.Close()
	if err := write(f); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, path); err != nil {
		return err
	}
	return syncDirectory(dir)
}

// MutableFile is geometric file-backed storage for one serial build.
type MutableFile struct {
	file *os.File
	data []byte
	used int
}

func NewMutableFile(f *os.File) *MutableFile { return &MutableFile{file: f} }
func (m *MutableFile) Append(b []byte) ([]byte, error) {
	size := m.used + len(b)
	if size > len(m.data) {
		capacity := max(size, len(m.data)+len(m.data)/2, 4096)
		if m.data != nil {
			if err := unmapWritable(m.file, m.data, m.used); err != nil {
				return nil, err
			}
			m.data = nil
		}
		if err := m.file.Truncate(int64(capacity)); err != nil {
			return nil, err
		}
		data, err := mapWritable(m.file, capacity)
		if err != nil {
			return nil, err
		}
		m.data = data
	}
	copy(m.data[m.used:size], b)
	m.used = size
	return m.data[:m.used], nil
}
func (m *MutableFile) Close() error {
	if m.data != nil {
		data := m.data
		m.data = nil
		if err := unmapWritable(m.file, data, m.used); err != nil {
			return err
		}
	}
	return m.file.Truncate(int64(m.used))
}
