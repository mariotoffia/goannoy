package dot

import (
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/engine"
)

// Searcher owns reusable scratch and must not be shared by concurrent callers.
type Searcher struct{ inner *engine.Searcher }

func (i *Index) NewSearcher() *Searcher { return &Searcher{inner: i.inner.NewSearcher()} }
func (i *Index) VectorInto(id goannoy.ItemID, dst []float32) error {
	return i.inner.VectorFloat(id, dst)
}
func (i *Index) Vector(id goannoy.ItemID) ([]float32, error) {
	v := make([]float32, i.Dimension())
	if err := i.VectorInto(id, v); err != nil {
		return nil, err
	}
	return v, nil
}
func (i *Index) Distance(a, b goannoy.ItemID) (float32, error) { return i.inner.Distance(a, b) }
func (s *Searcher) SearchInto(q []float32, dst []goannoy.Result, k goannoy.SearchK) (int, error) {
	return s.inner.SearchFloat(q, dst, k)
}
func (s *Searcher) Search(q []float32, limit int, k goannoy.SearchK) ([]goannoy.Result, error) {
	if limit < 0 {
		return nil, goannoy.ErrInvalidOption
	}
	dst := make([]goannoy.Result, limit)
	n, err := s.SearchInto(q, dst, k)
	if err != nil {
		return nil, err
	}
	return dst[:n], nil
}
func (s *Searcher) SearchByIDInto(id goannoy.ItemID, dst []goannoy.Result, k goannoy.SearchK) (int, error) {
	return s.inner.SearchID(id, dst, k)
}
