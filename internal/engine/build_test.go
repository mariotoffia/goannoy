package engine

import (
	"bytes"
	"context"
	"errors"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"os"
	"path/filepath"
	"testing"
)

func fixtureFloat(id int) []float32 {
	return []float32{float32((id*17)%97-48) / 8, float32((id*31)%89-44) / 8, float32((id*43)%83-41) / 8}
}
func fixtureWords(id uint64) []uint64 {
	return []uint64{id * 0x9e3779b97f4a7c15, (id + 3) * 0xd6e8feb86659fd93, id & 3}
}

func TestDeterministicBuild(t *testing.T) {
	for _, k := range []codec.Kind{codec.Angular, codec.Dot, codec.Euclidean, codec.Manhattan, codec.Hamming} {
		t.Run(k.String(), func(t *testing.T) {
			dim := 3
			if k == codec.Hamming {
				dim = 130
			}
			b, e := NewBuilder(k, dim)
			if e != nil {
				t.Fatal(e)
			}
			b.Config.Trees(3).Workers(1).Seed(42)
			for id := 0; id < 64; id++ {
				if k == codec.Hamming {
					e = b.AddWords(goannoy.ItemID(id), fixtureWords(uint64(id)))
				} else {
					e = b.AddFloat(goannoy.ItemID(id), fixtureFloat(id))
				}
				if e != nil {
					t.Fatal(e)
				}
			}
			path := filepath.Join(t.TempDir(), k.String()+".ann")
			i, e := b.BuildTo(context.Background(), path)
			if e != nil {
				t.Fatal(e)
			}
			defer i.Close()
			got, e := os.ReadFile(path)
			if e != nil {
				t.Fatal(e)
			}
			want, e := os.ReadFile("../../compat/fixtures/" + k.String() + ".ann")
			if e != nil {
				t.Fatal(e)
			}
			// C++ leaves split-only padding and unused fields uninitialized.
			// Compare every defined field and byte; never bless a new golden.
			if len(got) == len(want) {
				l, _ := codec.New(k, dim)
				for id := 64; id < len(want)/l.Stride; id++ {
					if int(l.Descendants(want, id)) <= l.LeafCapacity {
						continue
					}
					a, z := l.Node(got, id), l.Node(want, id)
					if k == codec.Dot {
						clear(a[16:24])
						clear(z[16:24])
					}
					if k == codec.Hamming {
						clear(a[12:16])
						clear(z[12:16])
						clear(a[24:])
						clear(z[24:])
					}
				}
			}
			if !bytes.Equal(got, want) {
				n := min(len(got), len(want))
				at := 0
				for at < n && got[at] == want[at] {
					at++
				}
				t.Fatalf("bytes mismatch at %d: sizes %d/%d; got %x want %x", at, len(got), len(want), got[at:min(at+12, len(got))], want[at:min(at+12, len(want))])
			}
		})
	}
}

func TestBuilderSnapshotsAndInputs(t *testing.T) {
	b, e := NewBuilder(codec.Angular, 2)
	if e != nil {
		t.Fatal(e)
	}
	b.Config.Trees(2).Workers(1)
	if e = b.AddFloat(-1, []float32{1, 0}); !errors.Is(e, goannoy.ErrInvalidID) {
		t.Fatal(e)
	}
	if e = b.AddFloat(0, []float32{1}); !errors.Is(e, goannoy.ErrInvalidVector) {
		t.Fatal(e)
	}
	v := []float32{1, 0}
	if e = b.AddFloat(0, v); e != nil {
		t.Fatal(e)
	}
	v[0] = 42
	i, e := b.Build(nil)
	if e != nil {
		t.Fatal(e)
	}
	defer i.Close()
	dst := make([]float32, 2)
	if e = i.VectorFloat(0, dst); e != nil || dst[0] != 1 {
		t.Fatalf("%v %v", dst, e)
	}
	if e = b.AddFloat(0, []float32{0, 1}); e != nil {
		t.Fatal(e)
	}
	if e = b.AddFloat(3, []float32{-1, 0}); e != nil {
		t.Fatal(e)
	}
	j, e := b.Build(nil)
	if e != nil {
		t.Fatal(e)
	}
	defer j.Close()
	if j.Items() != 4 || j.Trees() != 2 {
		t.Fatalf("items=%d trees=%d", j.Items(), j.Trees())
	}
	if e = j.VectorFloat(1, dst); !errors.Is(e, goannoy.ErrInvalidID) {
		t.Fatal(e)
	}
	if e = i.VectorFloat(0, dst); e != nil || dst[0] != 1 {
		t.Fatal("old snapshot changed")
	}
	b.Config.Trees(0)
	if _, e = b.Build(nil); !errors.Is(e, goannoy.ErrInvalidOption) {
		t.Fatal(e)
	}
}
