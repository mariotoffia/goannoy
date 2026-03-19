//go:build !(linux || darwin || freebsd || openbsd || netbsd)

package memory

func (mi *mmapIndexAllocation) Prefault() error {
	return nil
}
