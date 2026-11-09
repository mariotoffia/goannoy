package engine

import (
	"context"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/core"
	"github.com/mariotoffia/goannoy/internal/metric"
	"math"
)

type treeBuilder struct {
	layout  codec.Layout
	data    []byte
	items   int
	ctx     context.Context
	random  *core.Kiss64
	storage *core.MutableFile
	spec    Spec
}

func (s *treeBuilder) appendNode(n []byte) (int32, error) {
	id := int32(len(s.data) / s.layout.Stride)
	if s.storage != nil {
		data, err := s.storage.Append(n)
		if err != nil {
			return 0, err
		}
		s.data = data
		return id, nil
	}
	s.data = append(s.data, n...)
	return id, nil
}
func imbalance(a, b int) float64 {
	f := float32(float64(float32(a)) / (float64(float32(a)) + float64(float32(b)) + 1e-9))
	return float64(max(f, 1-f))
}
func (s *treeBuilder) tree(ids []int32, root bool) (int32, error) {
	if err := s.ctx.Err(); err != nil {
		return 0, err
	}
	l := s.layout
	if len(ids) == 1 && !root {
		return ids[0], nil
	}
	ndesc := len(ids)
	if root {
		ndesc = s.items
	}
	plane := make([]byte, l.Stride)
	if len(ids) <= l.LeafCapacity && (!root || s.items <= l.LeafCapacity || len(ids) == 1) {
		codec.PutI32(plane, 0, int32(ndesc))
		for j, id := range ids {
			codec.PutI32(plane, l.ChildrenOffset+4*j, id)
		}
		return s.appendNode(plane)
	}
	sides := [2][]int32{make([]int32, 0, len(ids)), make([]int32, 0, len(ids))}
	for attempt := 0; attempt < 3; attempt++ {
		sides[0] = sides[0][:0]
		sides[1] = sides[1][:0]
		s.spec.split(l, s.data, ids, s.random, plane)
		for _, id := range ids {
			margin := metric.BuildMargin(l, plane, l.Node(s.data, int(id)))
			side := 0
			if l.Kind == codec.Hamming {
				side = int(margin)
			} else if margin == 0 {
				side = s.random.Flip()
			} else if margin > 0 {
				side = 1
			}
			sides[side] = append(sides[side], id)
		}
		if imbalance(len(sides[0]), len(sides[1])) < .95 {
			break
		}
	}
	for imbalance(len(sides[0]), len(sides[1])) > .99 {
		if err := s.ctx.Err(); err != nil {
			return 0, err
		}
		sides[0] = sides[0][:0]
		sides[1] = sides[1][:0]
		if l.Kind == codec.Hamming {
			clear(l.WordsAt(plane, 0))
		} else {
			clear(l.Floats(plane, 0))
		}
		for _, id := range ids {
			side := s.random.Flip()
			sides[side] = append(sides[side], id)
		}
	}
	codec.PutI32(plane, 0, int32(ndesc))
	flip := 0
	if len(sides[0]) > len(sides[1]) {
		flip = 1
	}
	for pass := 0; pass < 2; pass++ {
		side := pass ^ flip
		child, err := s.tree(sides[side], false)
		if err != nil {
			return 0, err
		}
		codec.PutI32(plane, l.ChildrenOffset+4*side, child)
	}
	if len(s.data)/l.Stride >= math.MaxInt32 {
		return 0, goannoy.ErrInvalidOption
	}
	return s.appendNode(plane)
}
