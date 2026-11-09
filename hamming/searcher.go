package hamming

import (
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/engine"
)

// Searcher owns reusable scratch and must not be shared by concurrent callers.
type Searcher struct{ inner *engine.Searcher }

func (i *Index) NewSearcher() *Searcher { return &Searcher{inner: i.inner.NewSearcher()} }
func (i *Index) VectorInto(id goannoy.ItemID, dst []uint64) error {
	return i.inner.VectorWords(id, dst)
}
func (i *Index) Vector(id goannoy.ItemID) ([]uint64, error) {
	v := make([]uint64, (i.Dimension()+63)/64)
	if err := i.VectorInto(id, v); err != nil {
		return nil, err
	}
	return v, nil
}
func (i *Index) Distance(a, b goannoy.ItemID) (float32, error) { return i.inner.Distance(a, b) }
func (s *Searcher) SearchInto(q []uint64, dst []goannoy.Result, k goannoy.SearchK) (int, error) {
	return s.inner.SearchWords(q, dst, k)
}
func (s *Searcher) Search(q []uint64, limit int, k goannoy.SearchK) ([]goannoy.Result, error) {
	dst, k, err := s.inner.PrepareResults(limit, k)
	if err != nil {
		return nil, err
	}
	n, err := s.SearchInto(q, dst, k)
	if err != nil {
		return nil, err
	}
	return dst[:n], nil
}
func (s *Searcher) SearchByIDInto(id goannoy.ItemID, dst []goannoy.Result, k goannoy.SearchK) (int, error) {
	return s.inner.SearchID(id, dst, k)
}
