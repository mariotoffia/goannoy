package core

import (
	"encoding/binary"
	"errors"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"os"
	"testing"
)

func TestOpenSpotifyFixtures(t *testing.T) {
	for _, k := range []codec.Kind{codec.Angular, codec.Dot, codec.Euclidean, codec.Manhattan, codec.Hamming} {
		t.Run(k.String(), func(t *testing.T) {
			dim := 3
			if k == codec.Hamming {
				dim = 130
			}
			l, _ := codec.New(k, dim)
			b, err := os.ReadFile("../../compat/fixtures/" + k.String() + ".ann")
			if err != nil {
				t.Fatal(err)
			}
			info, err := Validate(b, l)
			if err != nil {
				t.Fatal(err)
			}
			if info.Items != 64 || len(info.Roots) != 3 {
				t.Fatalf("%+v", info)
			}
		})
	}
}

// Catches unsafe traversal before data is exposed to search.
func TestOpenRejectsMalformedIndex(t *testing.T) {
	l, _ := codec.New(codec.Angular, 3)
	good, err := os.ReadFile("../../compat/fixtures/angular.ann")
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string][]byte{"empty": nil, "short": {1}, "misaligned": good[:len(good)-1]}
	for _, tc := range []struct {
		name string
		off  int
		v    uint32
	}{
		{"negative descendants", 0, 0xffffffff}, {"oversized root", len(good) - l.Stride, 0x7fffffff},
		{"child beyond file", 64*l.Stride + 4, 0x7fffffff}, {"cycle", len(good) - l.Stride + 4, uint32(len(good)/l.Stride - 1)},
	} {
		b := append([]byte(nil), good...)
		binary.LittleEndian.PutUint32(b[tc.off:], tc.v)
		cases[tc.name] = b
	}
	for name, b := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Validate(b, l); !errors.Is(err, goannoy.ErrCorruptIndex) {
				t.Fatalf("accepted malformed file: %v", err)
			}
		})
	}
}

func FuzzOpenIndex(f *testing.F) {
	b, err := os.ReadFile("../../compat/fixtures/angular.ann")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(b)
	f.Add([]byte{})
	f.Add([]byte{1, 0, 0, 0})
	l, _ := codec.New(codec.Angular, 3)
	f.Fuzz(func(t *testing.T, b []byte) { _, _ = Validate(b, l) })
}
