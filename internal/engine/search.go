package engine

import (
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/core"
	"github.com/mariotoffia/goannoy/internal/metric"
	"math"
	"unsafe"
)

// Searcher is scratch owned by one caller; Index remains immutable.
type Searcher struct {
	index *Index
	queue core.CandidateHeap
	ids   []int32
	query []byte
}

var positiveInfinity = float32(math.Inf(1))

// PrepareResults bounds allocation while preserving the original AutoSearchK
// budget when the requested result count is larger than the index.
func (s *Searcher) PrepareResults(limit int, k goannoy.SearchK) ([]goannoy.Result, goannoy.SearchK, error) {
	if limit < 0 || limit > math.MaxInt/int(unsafe.Sizeof(goannoy.Result{})) {
		return nil, k, goannoy.ErrInvalidOption
	}
	if k == goannoy.AutoSearchK && limit > 0 {
		trees := len(s.index.info.Roots)
		if limit > math.MaxInt/trees {
			return nil, k, goannoy.ErrInvalidSearchK
		}
		k = goannoy.SearchK(limit * trees)
	}
	return make([]goannoy.Result, min(limit, s.index.info.Items)), k, nil
}

func finiteDistance(x float32) bool { return x <= math.MaxFloat32 && x >= -math.MaxFloat32 }

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
	if k < goannoy.AutoSearchK || k == 0 {
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
	return i.spec.search(s, dst, budget)
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
