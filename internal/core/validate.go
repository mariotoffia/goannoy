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
	if items <= l.LeafCapacity && len(info.Roots)%2 == 0 {
		half := len(info.Roots) / 2
		if bytes.Equal(data[(count-half*2)*l.Stride:(count-half)*l.Stride], data[(count-half)*l.Stride:]) {
			info.Roots = info.Roots[:half]
		}
	} else if len(info.Roots) > 1 {
		first, last := info.Roots[0], info.Roots[len(info.Roots)-1]
		if l.Children(data, int(first))[0] == l.Children(data, int(last))[0] {
			info.Roots = info.Roots[:len(info.Roots)-1]
		}
	}
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
				if l.Kind == codec.Hamming {
					if l.CheckWords(l.WordsAt(data, id)) != nil {
						return bad("invalid hamming item")
					}
				} else {
					if l.CheckFloats(l.Floats(data, id)) != nil {
						return bad("non-finite item")
					}
				}
			}
			continue
		}
		if n < 1 {
			return bad("empty tree node")
		}
		if n <= l.LeafCapacity {
			off := id*l.Stride + l.ChildrenOffset
			for j := 0; j < n; j++ {
				child := int(codec.I32(data, off+4*j))
				if child < 0 || child >= items || l.Descendants(data, child) != 1 {
					return bad("invalid leaf ID")
				}
			}
		} else {
			for _, child := range l.Children(data, id) {
				if child < 0 || int(child) >= id {
					return bad("invalid or cyclic child edge")
				}
			}
			if l.Kind == codec.Hamming {
				if l.WordsAt(data, id)[0] >= uint64(l.Words*64) {
					return bad("hamming split coordinate out of range")
				}
			} else if l.CheckFloats(l.Floats(data, id)) != nil {
				return bad("non-finite split")
			}
		}
	}
	return info, nil
}
