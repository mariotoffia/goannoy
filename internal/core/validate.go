package core

import (
	"bytes"
	"fmt"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"math"
)

type IndexInfo struct {
	Items int
	Roots []int32
}

// Validate checks every offset and edge before traversal. Standard Annoy files
// contain no metric/dimension header; callers must retain that metadata.
func Validate(data []byte, l codec.Layout) (IndexInfo, error) {
	bad := func(reason string) (IndexInfo, error) {
		return IndexInfo{}, fmt.Errorf("%w: %s", goannoy.ErrCorruptIndex, reason)
	}
	finite := func(x float32) bool { return !math.IsNaN(float64(x)) && !math.IsInf(float64(x), 0) }
	if l.Stride < 1 || len(data) == 0 || len(data)%l.Stride != 0 {
		return bad("invalid file length")
	}
	count := len(data) / l.Stride
	if count > math.MaxInt32 {
		return bad("too many nodes")
	}
	items := int(l.Descendants(data, count-1))
	if items < 1 || items >= count {
		return bad("invalid item count")
	}
	info := IndexInfo{Items: items}
	for n := count - 1; n >= items && int(l.Descendants(data, n)) == items; n-- {
		info.Roots = append(info.Roots, int32(n))
	}
	if len(info.Roots) == 0 {
		return bad("missing roots")
	}
	// Tiny trees occupy one list node each and the entire root array is copied.
	duplicatedBlock := false
	if len(info.Roots)%2 == 0 {
		half := len(info.Roots) / 2
		if bytes.Equal(data[(count-half*2)*l.Stride:(count-half)*l.Stride], data[(count-half)*l.Stride:]) {
			info.Roots = info.Roots[:half]
			duplicatedBlock = true
		}
	}
	if !duplicatedBlock && len(info.Roots) > 1 {
		first, last := info.Roots[0], info.Roots[len(info.Roots)-1]
		if l.Children(data, int(first))[0] == l.Children(data, int(last))[0] {
			info.Roots = info.Roots[:len(info.Roots)-1]
		}
	}
	present := 0
	for id := 0; id < count; id++ {
		n := int(l.Descendants(data, id))
		if n < 0 || n > items {
			return bad(fmt.Sprintf("node %d descendant count", id))
		}
		if id < items {
			if n != 0 && n != 1 {
				return bad("item node is not a leaf")
			}
			if n == 1 {
				present++
				if l.Kind == codec.Hamming {
					if l.CheckWords(l.WordsAt(data, id)) != nil {
						return bad("invalid hamming item")
					}
				} else {
					if l.CheckFloats(l.Floats(data, id)) != nil {
						return bad("non-finite item")
					}
					node := l.Node(data, id)
					if l.Kind == codec.Angular && (!finite(codec.F32(node, 4)) || codec.F32(node, 4) < 0) {
						return bad("invalid angular norm")
					}
					if l.Kind == codec.Dot && (!finite(codec.F32(node, 12)) || codec.F32(node, 12) < 0 || !finite(codec.F32(node, 16)) || codec.F32(node, 16) < 0 || node[20] != 1) {
						return bad("invalid dot item metadata")
					}
				}
			}
			continue
		}
		if n < 1 || present == 0 {
			return bad("empty tree node")
		}
		if n <= l.LeafCapacity {
			off := id*l.Stride + l.ChildrenOffset
			for j := 0; j < n; j++ {
				child := int(codec.I32(data, off+4*j))
				// Compact sparse roots store maxID+1 as n_descendants but
				// contain only present IDs followed by zero padding.
				if n == items && j >= present && child == 0 {
					continue
				}
				if child < 0 || child >= items || l.Descendants(data, child) != 1 {
					return bad("invalid leaf ID")
				}
			}
		} else {
			children := l.Children(data, id)
			if children[0] == children[1] {
				return bad("identical split children")
			}
			total := 0
			for _, child := range children {
				if child < 0 || int(child) >= id {
					return bad("invalid or cyclic child edge")
				}
				cn := int(l.Descendants(data, int(child)))
				// Upstream encodes a sparse single-item root like a split;
				// its unused second child is the empty item-zero slot.
				padding := n == items && present == 1 && child == 0
				if cn >= n || (cn == 0 && !padding) {
					return bad("invalid child descendant count")
				}
				total += cn
			}
			want := n
			if n == items {
				want = present
			}
			if total != want {
				return bad("split descendant total mismatch")
			}
			if l.Kind == codec.Hamming {
				if l.WordsAt(data, id)[0] >= uint64(l.Words*64) {
					return bad("hamming split coordinate out of range")
				}
			} else if l.CheckFloats(l.Floats(data, id)) != nil {
				return bad("non-finite split")
			}
			node := l.Node(data, id)
			if l.Kind == codec.Dot && !finite(codec.F32(node, 12)) {
				return bad("invalid dot plane factor")
			}
			if (l.Kind == codec.Euclidean || l.Kind == codec.Manhattan) && !finite(codec.F32(node, 4)) {
				return bad("invalid plane offset")
			}
		}
	}
	return info, nil
}
