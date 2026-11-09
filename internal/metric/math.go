// Package metric contains scalar Spotify Annoy distance and split primitives.
package metric

import (
	"github.com/mariotoffia/goannoy/internal/codec"
	"math"
	"math/bits"
)

func Dot(a, b []float32) float32 {
	var v float32
	for n, x := range a {
		v += float32(x * b[n])
	}
	return v
}
func Squared(a, b []float32) float32 {
	var v float32
	for n, x := range a {
		d := x - b[n]
		v += float32(d * d)
	}
	return v
}
func L1(a, b []float32) float32 {
	var v float32
	for n, x := range a {
		d := x - b[n]
		if d < 0 {
			d = -d
		}
		v += d
	}
	return v
}
func sqrt(x float32) float32 { return float32(math.Sqrt(float64(x))) }
func Norm(l codec.Layout, b []byte) float32 {
	v := l.Floats(b, 0)
	n := Dot(v, v)
	if l.Kind == codec.Dot {
		f := codec.F32(b, 12)
		n += float32(f * f)
	}
	return sqrt(n)
}
func Init(l codec.Layout, b []byte) {
	if l.Kind == codec.Angular {
		codec.PutF32(b, 4, Dot(l.Floats(b, 0), l.Floats(b, 0)))
	}
	if l.Kind == codec.Dot {
		f := codec.F32(b, 12)
		codec.PutF32(b, 16, Dot(l.Floats(b, 0), l.Floats(b, 0))+float32(f*f))
		b[20] = 0
	}
}
func Raw(l codec.Layout, a, b []byte) float32 {
	if l.Kind == codec.Hamming {
		x, y := l.WordsAt(a, 0), l.WordsAt(b, 0)
		d := 0
		for n, v := range x {
			d += bits.OnesCount64(v ^ y[n])
		}
		return float32(d)
	}
	x, y := l.Floats(a, 0), l.Floats(b, 0)
	switch l.Kind {
	case codec.Euclidean:
		return Squared(x, y)
	case codec.Manhattan:
		return L1(x, y)
	case codec.Dot:
		if a[20] != 0 || b[20] != 0 {
			return -Dot(x, y)
		}
	}
	off := 4
	if l.Kind == codec.Dot {
		off = 16
	}
	pp, qq, pq := codec.F32(a, off), codec.F32(b, off), Dot(x, y)
	if pp == 0 {
		pp = Dot(x, x)
	}
	if qq == 0 {
		qq = Dot(y, y)
	}
	if l.Kind == codec.Dot {
		af, bf := codec.F32(a, 12), codec.F32(b, 12)
		if codec.F32(a, off) == 0 {
			pp += float32(af * af)
		}
		if codec.F32(b, off) == 0 {
			qq += float32(bf * bf)
		}
		pq += float32(af * bf)
	}
	p := float32(pp * qq)
	if p > 0 {
		return float32(2 - 2*float64(pq)/math.Sqrt(float64(p)))
	}
	return 2
}
func Normalize(k codec.Kind, d float32) float32 {
	switch k {
	case codec.Dot:
		return -d
	case codec.Angular, codec.Euclidean:
		return sqrt(max(d, 0))
	default:
		return max(d, 0)
	}
}
func BuildMargin(l codec.Layout, plane, item []byte) float32 {
	if l.Kind == codec.Hamming {
		bit := l.WordsAt(plane, 0)[0]
		if l.WordsAt(item, 0)[bit/64]&(uint64(1)<<(63-bit%64)) != 0 {
			return 1
		}
		return 0
	}
	m := Dot(l.Floats(plane, 0), l.Floats(item, 0))
	if l.Kind == codec.Dot {
		m += float32(codec.F32(plane, 12) * codec.F32(item, 12))
	}
	if l.Kind == codec.Euclidean || l.Kind == codec.Manhattan {
		m += codec.F32(plane, 4)
	}
	return m
}
