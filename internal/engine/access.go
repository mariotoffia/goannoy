package engine

import (
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/metric"
)

func (i *Index) validID(id goannoy.ItemID) bool {
	return id >= 0 && int(id) < i.info.Items && i.layout.Descendants(i.mapping.Data, int(id)) == 1
}
func (i *Index) VectorFloat(id goannoy.ItemID, dst []float32) error {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.closed {
		return goannoy.ErrClosed
	}
	if !i.validID(id) {
		return goannoy.ErrInvalidID
	}
	if len(dst) != i.layout.Dimension {
		return goannoy.ErrInvalidVector
	}
	copy(dst, i.layout.Floats(i.mapping.Data, int(id)))
	return nil
}
func (i *Index) VectorWords(id goannoy.ItemID, dst []uint64) error {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.closed {
		return goannoy.ErrClosed
	}
	if !i.validID(id) {
		return goannoy.ErrInvalidID
	}
	if len(dst) != i.layout.Words {
		return goannoy.ErrInvalidVector
	}
	copy(dst, i.layout.WordsAt(i.mapping.Data, int(id)))
	return nil
}
func (i *Index) Distance(a, b goannoy.ItemID) (float32, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.closed {
		return 0, goannoy.ErrClosed
	}
	if !i.validID(a) || !i.validID(b) {
		return 0, goannoy.ErrInvalidID
	}
	l, data := i.layout, i.mapping.Data
	return metric.Normalize(l.Kind, metric.Raw(l, l.Node(data, int(a)), l.Node(data, int(b)))), nil
}
