package metric

import (
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/core"
	"math"
)

func normalize(l codec.Layout, node []byte) {
	norm := Norm(l, node)
	if l.Kind == codec.Dot {
		f := float64(codec.F32(node, 12))
		norm = float32(math.Sqrt(float64(Dot(l.Floats(node, 0), l.Floats(node, 0))) + f*f))
	}
	if norm > 0 {
		for i := range l.Floats(node, 0) {
			l.Floats(node, 0)[i] /= norm
		}
		if l.Kind == codec.Dot {
			codec.PutF32(node, 12, codec.F32(node, 12)/norm)
		}
	}
}

func Split(l codec.Layout, data []byte, ids []int32, r *core.Kiss64, plane []byte) {
	if l.Kind == codec.Hamming {
		splitHamming(l, data, ids, r, plane)
		return
	}
	scratch := make([]byte, l.Stride*2)
	p, q := scratch[:l.Stride], scratch[l.Stride:]
	a, b := r.Index(len(ids)), r.Index(len(ids)-1)
	if b >= a {
		b++
	}
	copy(l.Floats(p, 0), l.Floats(data, int(ids[a])))
	copy(l.Floats(q, 0), l.Floats(data, int(ids[b])))
	if l.Kind == codec.Dot {
		codec.PutF32(p, 12, codec.F32(l.Node(data, int(ids[a])), 12))
		codec.PutF32(q, 12, codec.F32(l.Node(data, int(ids[b])), 12))
	}
	cosine := l.Kind == codec.Angular || l.Kind == codec.Dot
	if cosine {
		normalize(l, p)
		normalize(l, q)
	}
	Init(l, p)
	Init(l, q)
	ic, jc := 1, 1
	for step := 0; step < 200; step++ {
		item := l.Node(data, int(ids[r.Index(len(ids))]))
		di, dj := float32(ic)*Raw(l, p, item), float32(jc)*Raw(l, q, item)
		norm := float32(1)
		if cosine {
			norm = Norm(l, item)
		}
		if !(norm > 0) {
			continue
		}
		if di < dj {
			updateMean(l, p, item, norm, ic)
			Init(l, p)
			ic++
		} else if dj < di {
			updateMean(l, q, item, norm, jc)
			Init(l, q)
			jc++
		}
	}
	pv, qv, v := l.Floats(p, 0), l.Floats(q, 0), l.Floats(plane, 0)
	for i := range v {
		v[i] = pv[i] - qv[i]
	}
	if l.Kind == codec.Dot {
		codec.PutF32(plane, 12, codec.F32(p, 12)-codec.F32(q, 12))
	}
	normalize(l, plane)
	if l.Kind == codec.Euclidean || l.Kind == codec.Manhattan {
		var a float32
		for i, x := range v {
			a += float32(-x*(pv[i]+qv[i])) / 2
		}
		codec.PutF32(plane, 4, a)
	}
}
func updateMean(l codec.Layout, mean, item []byte, norm float32, count int) {
	m, v := l.Floats(mean, 0), l.Floats(item, 0)
	c, d := float32(count), float32(count+1)
	for i := range m {
		m[i] = (float32(m[i]*c) + v[i]/norm) / d
	}
	if l.Kind == codec.Dot {
		codec.PutF32(mean, 12, (float32(codec.F32(mean, 12)*c)+codec.F32(item, 12)/norm)/d)
	}
}
func splitHamming(l codec.Layout, data []byte, ids []int32, r *core.Kiss64, plane []byte) {
	dim := l.Words * 64
	partition := func(bit int) bool {
		l.WordsAt(plane, 0)[0] = uint64(bit)
		count := 0
		for _, id := range ids {
			count += int(BuildMargin(l, plane, l.Node(data, int(id))))
		}
		return count > 0 && count < len(ids)
	}
	for i := 0; i < 20; i++ {
		if partition(r.Index(dim)) {
			return
		}
	}
	for i := 0; i < dim; i++ {
		if partition(i) {
			return
		}
	}
}
