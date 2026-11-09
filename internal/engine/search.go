package engine

import (
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/core"
	"github.com/mariotoffia/goannoy/internal/metric"
	"math"
)

// Searcher is scratch owned by one caller; Index remains immutable.
type Searcher struct {
	index *Index
	queue core.CandidateHeap
	ids   []int32
	query []byte
}

var positiveInfinity = float32(math.Inf(1))

func (i *Index) NewSearcher() *Searcher {
	return &Searcher{index: i, query: make([]byte, i.layout.Stride)}
}
func (s *Searcher) SearchFloat(q []float32, dst []goannoy.Result, k goannoy.SearchK) (int, error) {
	i := s.index
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.closed {
		return 0, goannoy.ErrClosed
	}
	if e := i.layout.CheckFloats(q); e != nil {
		return 0, e
	}
	copy(i.layout.Floats(s.query, 0), q)
	metric.Init(i.layout, s.query)
	return s.search(dst, k)
}
func (s *Searcher) SearchWords(q []uint64, dst []goannoy.Result, k goannoy.SearchK) (int, error) {
	i := s.index
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.closed {
		return 0, goannoy.ErrClosed
	}
	if e := i.layout.CheckWords(q); e != nil {
		return 0, e
	}
	copy(i.layout.WordsAt(s.query, 0), q)
	return s.search(dst, k)
}
func (s *Searcher) SearchID(id goannoy.ItemID, dst []goannoy.Result, k goannoy.SearchK) (int, error) {
	i := s.index
	i.mu.RLock()
	defer i.mu.RUnlock()
	if i.closed {
		return 0, goannoy.ErrClosed
	}
	if !i.validID(id) {
		return 0, goannoy.ErrInvalidID
	}
	copy(s.query, i.layout.Node(i.mapping.Data, int(id)))
	// A dot query uses no lifted extra dimension in traversal.
	if i.layout.Kind == codec.Dot {
		codec.PutF32(s.query, 12, 0)
		metric.Init(i.layout, s.query)
	}
	return s.search(dst, k)
}

//go:generate go run ../genengine

func (s *Searcher) search(dst []goannoy.Result, k goannoy.SearchK) (int, error) {
	i := s.index
	if k < goannoy.AutoSearchK {
		return 0, goannoy.ErrInvalidSearchK
	}
	if len(dst) == 0 {
		return 0, nil
	}
	budget := int(k)
	if k == goannoy.AutoSearchK {
		if len(dst) > math.MaxInt/len(i.info.Roots) {
			return 0, goannoy.ErrInvalidSearchK
		}
		budget = len(dst) * len(i.info.Roots)
	}
	switch i.layout.Kind {
	case codec.Angular:
		return s.searchAngular(dst, budget)
	case codec.Dot:
		return s.searchDot(dst, budget)
	case codec.Euclidean:
		return s.searchEuclidean(dst, budget)
	case codec.Manhattan:
		return s.searchManhattan(dst, budget)
	default:
		return s.searchHamming(dst, budget)
	}
}
func compareResult(a, b goannoy.Result) int {
	if core.Worse(a, b) {
		return 1
	}
	if core.Worse(b, a) {
		return -1
	}
	return 0
}
func angularRaw(pq, pp, qq float32) float32 {
	product := float32(pp * qq)
	if product > 0 {
		return float32(2 - 2*float64(pq)/math.Sqrt(float64(product)))
	}
	return 2
}
