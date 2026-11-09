// Package codec describes Spotify Annoy's headerless standard node layouts.
package codec

import (
	"encoding/binary"
	"fmt"
	goannoy "github.com/mariotoffia/goannoy"
	"math"
	"unsafe"
)

type Kind uint8

const (
	Angular Kind = iota
	Dot
	Euclidean
	Manhattan
	Hamming
)

func (k Kind) String() string {
	switch k {
	case Angular:
		return "angular"
	case Dot:
		return "dot"
	case Euclidean:
		return "euclidean"
	case Manhattan:
		return "manhattan"
	case Hamming:
		return "hamming"
	}
	return "unknown"
}

type Layout struct {
	Kind                                                                 Kind
	Dimension, Words, Stride, VectorOffset, ChildrenOffset, LeafCapacity int
}

func New(k Kind, dimension int) (Layout, error) {
	var endian uint32 = 1
	if *(*byte)(unsafe.Pointer(&endian)) != 1 {
		return Layout{}, goannoy.ErrUnsupportedPlatform
	}
	if dimension < 1 || dimension > (math.MaxInt32-24)/8 {
		return Layout{}, fmt.Errorf("%w: %d", goannoy.ErrInvalidDimension, dimension)
	}
	// Result distances are float32; every supported Hamming count must be exact.
	if k == Hamming && dimension > 1<<24 {
		return Layout{}, goannoy.ErrInvalidDimension
	}
	l := Layout{Kind: k, Dimension: dimension, ChildrenOffset: 4}
	switch k {
	case Angular:
		l.VectorOffset = 12
	case Dot:
		l.VectorOffset = 24
	case Euclidean, Manhattan:
		l.VectorOffset = 16
		l.ChildrenOffset = 8
	case Hamming:
		l.VectorOffset = 16
		l.Words = (dimension + 63) / 64
	default:
		return Layout{}, goannoy.ErrInvalidOption
	}
	if k == Hamming {
		l.Stride = 16 + 8*l.Words
	} else {
		l.Stride = l.VectorOffset + 4*dimension
	}
	l.LeafCapacity = (l.Stride - l.ChildrenOffset) / 4
	return l, nil
}
func I32(b []byte, off int) int32       { return int32(binary.LittleEndian.Uint32(b[off:])) }
func PutI32(b []byte, off int, v int32) { binary.LittleEndian.PutUint32(b[off:], uint32(v)) }

// Float fields share the native, little-endian representation of vector views.
// Direct field loads also avoid checkptr=2 heap promotion in math bit casts.
func F32(b []byte, off int) float32 { return *(*float32)(unsafe.Pointer(&b[off : off+4][0])) }
func PutF32(b []byte, off int, v float32) {
	*(*float32)(unsafe.Pointer(&b[off : off+4][0])) = v
}
func (l Layout) Node(b []byte, id int) []byte       { return b[id*l.Stride : (id+1)*l.Stride] }
func (l Layout) Descendants(b []byte, id int) int32 { return I32(b, id*l.Stride) }
func (l Layout) Children(b []byte, id int) [2]int32 {
	off := id*l.Stride + l.ChildrenOffset
	return [2]int32{I32(b, off), I32(b, off+4)}
}

// Floats and WordsAt operate only on a complete, aligned, validated node.
// unsafe.Slice spans precisely the vector, never n_descendants elements.
func (l Layout) Floats(b []byte, id int) []float32 {
	node := l.Node(b, id)
	return unsafe.Slice((*float32)(unsafe.Pointer(&node[l.VectorOffset])), l.Dimension)
}
func (l Layout) WordsAt(b []byte, id int) []uint64 {
	node := l.Node(b, id)
	return unsafe.Slice((*uint64)(unsafe.Pointer(&node[l.VectorOffset])), l.Words)
}
func (l Layout) CheckWords(v []uint64) error {
	if len(v) != l.Words {
		return goannoy.ErrInvalidVector
	}
	if n := l.Dimension % 64; n != 0 && v[len(v)-1]>>n != 0 {
		return goannoy.ErrInvalidVector
	}
	return nil
}
func (l Layout) CheckFloats(v []float32) error {
	if len(v) != l.Dimension {
		return goannoy.ErrInvalidVector
	}
	var norm float32
	for _, x := range v {
		// Ordered comparisons reject NaN/Inf without math.Abs bit casts,
		// which checkptr=2 heap-promotes on some AMD64 toolchains.
		if !(x >= -math.MaxFloat32 && x <= math.MaxFloat32) {
			return goannoy.ErrInvalidVector
		}
		norm += float32(x * x)
	}
	if norm > math.MaxFloat32 {
		return goannoy.ErrInvalidVector
	}
	return nil
}
