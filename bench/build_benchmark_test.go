package bench

import (
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/angular"
	"testing"
)

func BenchmarkAngularBuild(b *testing.B) {
	v := vectors(10000, 128, 42)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		builder, e := angular.NewBuilder(128)
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
		if e := i.Close(); e != nil {
			b.Fatal(e)
		}
	}
}
