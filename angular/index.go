// Package angular implements Spotify Annoy's angular metric.
package angular

import (
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/engine"
)

// Index is immutable and safe for concurrent use. Close waits for active queries.
type Index struct{ inner *engine.Index }

// Open validates and loads a standard headerless Annoy file.
func Open(path string, dimension int, options ...goannoy.OpenOption) (*Index, error) {
	i, err := engine.Open(path, codec.Angular, dimension, options...)
	if err != nil {
		return nil, err
	}
	return &Index{inner: i}, nil
}
func (i *Index) Dimension() int { return i.inner.Dimension() }
func (i *Index) Items() int     { return i.inner.Items() }
func (i *Index) Trees() int     { return i.inner.Trees() }
func (i *Index) Close() error   { return i.inner.Close() }
