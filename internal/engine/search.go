package engine

import (
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/core"
	"github.com/mariotoffia/goannoy/internal/metric"
	"math"
	"slices"
)

// Searcher is scratch owned by one caller; Index remains immutable.
type Searcher struct {
	index *Index
	queue core.CandidateHeap
	ids   []int32
	query []byte
}

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
func (s *Searcher) search(dst []goannoy.Result, k goannoy.SearchK) (int, error) {
	i := s.index
	l, data := i.layout, i.mapping.Data
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
	s.queue = s.queue[:0]
	s.ids = s.ids[:0]
	initial := float32(math.Inf(1))
	if l.Kind == codec.Hamming {
		// Translation by a constant preserves the integer penalty ordering.
		// Infinity (or rounded uint64 max) would swallow one-bit penalties.
		initial = 0
	}
	for _, root := range i.info.Roots {
		s.queue.Push(core.Candidate{Priority: initial, ID: root})
	}
	for len(s.ids) < budget && len(s.queue) > 0 {
		c := s.queue.Pop()
		id := int(c.ID)
		n := int(l.Descendants(data, id))
		if n == 1 && id < i.info.Items {
			s.ids = append(s.ids, c.ID)
		} else if n <= l.LeafCapacity {
			off := id*l.Stride + l.ChildrenOffset
			for j := 0; j < n; j++ {
				s.ids = append(s.ids, codec.I32(data, off+j*4))
			}
		} else {
			// Query dot factors are zero; other metrics use the same margin as build.
			margin := metric.BuildMargin(l, l.Node(data, id), s.query)
			children := l.Children(data, id)
			for side := 1; side >= 0; side-- {
				priority := c.Priority
				if l.Kind == codec.Hamming {
					if int(margin) != side {
						priority--
					}
				} else {
					signed := margin
					if side == 0 {
						signed = -signed
					}
					priority = min(priority, signed)
				}
				s.queue.Push(core.Candidate{Priority: priority, ID: children[side]})
			}
		}
	}
	slices.Sort(s.ids)
	n := 0
	last := int32(-1)
	for _, id := range s.ids {
		if id == last {
			continue
		}
		last = id
		if l.Descendants(data, int(id)) != 1 {
			continue
		}
		raw := metric.Raw(l, s.query, l.Node(data, int(id)))
		n = core.Offer(dst, n, goannoy.Result{ID: goannoy.ItemID(id), Distance: raw})
	}
	slices.SortFunc(dst[:n], func(a, b goannoy.Result) int {
		if core.Worse(a, b) {
			return 1
		}
		if core.Worse(b, a) {
			return -1
		}
		return 0
	})
	for j := 0; j < n; j++ {
		dst[j].Distance = metric.Normalize(l.Kind, dst[j].Distance)
	}
	return n, nil
}
