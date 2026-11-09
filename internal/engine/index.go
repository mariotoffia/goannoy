// Package engine owns immutable storage and the common operation lifecycle.
package engine

import (
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/core"
	"sync"
)

type Index struct {
	mu       sync.RWMutex
	layout   codec.Layout
	mapping  *core.Mapping
	info     core.IndexInfo
	closed   bool
	closeErr error
}

func Open(path string, kind codec.Kind, dimension int, options ...goannoy.OpenOption) (*Index, error) {
	l, err := codec.New(kind, dimension)
	if err != nil {
		return nil, err
	}
	m, err := core.OpenMapping(path, options...)
	if err != nil {
		return nil, err
	}
	i, err := FromMapping(m, l)
	if err != nil {
		_ = m.Close()
		return nil, err
	}
	return i, nil
}
func FromMapping(m *core.Mapping, l codec.Layout) (*Index, error) {
	info, err := core.Validate(m.Data, l)
	if err != nil {
		return nil, err
	}
	return &Index{layout: l, mapping: m, info: info}, nil
}
func (i *Index) Dimension() int { return i.layout.Dimension }
func (i *Index) Items() int     { return i.info.Items }
func (i *Index) Trees() int     { return len(i.info.Roots) }
func (i *Index) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.closed {
		return i.closeErr
	}
	i.closed = true
	i.closeErr = i.mapping.Close()
	return i.closeErr
}
