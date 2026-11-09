package bench

import (
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/angular"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/engine"
	"math/rand"
	"testing"
)

func vectors(items, dim int, seed int64) [][]float32 {
	r := rand.New(rand.NewSource(seed))
	v := make([][]float32, items)
	for i := range v {
		v[i] = make([]float32, dim)
		for j := range v[i] {
			v[i][j] = float32(r.NormFloat64())
		}
	}
	return v
}

func TestSearchIntoAllocs(t *testing.T) {
	for _, k := range []codec.Kind{codec.Angular, codec.Dot, codec.Euclidean, codec.Manhattan, codec.Hamming} {
		t.Run(k.String(), func(t *testing.T) {
			dim := 3
			if k == codec.Hamming {
				dim = 130
			}
			i, e := engine.Open("../compat/fixtures/"+k.String()+".ann", k, dim)
			if e != nil {
				t.Fatal(e)
			}
			defer i.Close()
			s := i.NewSearcher()
			dst := make([]goannoy.Result, 10)
			f := func() {
				var e error
				if k == codec.Hamming {
					_, e = s.SearchWords([]uint64{1, 2, 1}, dst, -1)
				} else {
					_, e = s.SearchFloat([]float32{1, 2, 3}, dst, -1)
				}
				if e != nil {
					panic(e)
				}
			}
			f()
			if n := testing.AllocsPerRun(100, f); n != 0 {
				t.Fatalf("%g allocs/op", n)
			}
		})
	}
}

func BenchmarkAngularSearch(b *testing.B) {
	for _, dim := range []int{128, 512, 1536} {
		b.Run(dimName(dim), func(b *testing.B) {
			v, q := vectors(10000, dim, 99), vectors(128, dim, 1234)
			builder, e := angular.NewBuilder(dim)
			if e != nil {
				b.Fatal(e)
			}
			builder.Trees(10).Workers(1).Seed(42).Capacity(10000)
			for id, x := range v {
				if e := builder.Add(goannoy.ItemID(id), x); e != nil {
					b.Fatal(e)
				}
			}
			i, e := builder.Build(nil)
			if e != nil {
				b.Fatal(e)
			}
			defer i.Close()
			s := i.NewSearcher()
			dst := make([]goannoy.Result, 10)
			for _, x := range q {
				if _, e := s.SearchInto(x, dst, -1); e != nil {
					b.Fatal(e)
				}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				if _, e := s.SearchInto(q[n%len(q)], dst, -1); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}
func dimName(d int) string {
	switch d {
	case 128:
		return "dim128"
	case 512:
		return "dim512"
	default:
		return "dim1536"
	}
}
