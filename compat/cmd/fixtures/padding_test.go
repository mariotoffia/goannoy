package main

import (
	"bytes"
	"testing"

	"github.com/mariotoffia/goannoy/internal/codec"
)

func TestCanonicalPaddingPreservesDefinedFields(t *testing.T) {
	for _, kind := range []codec.Kind{codec.Dot, codec.Hamming} {
		t.Run(kind.String(), func(t *testing.T) {
			dim := 3
			if kind == codec.Hamming {
				dim = 130
			}
			l, err := codec.New(kind, dim)
			if err != nil {
				t.Fatal(err)
			}
			data := bytes.Repeat([]byte{0xa5}, l.Stride*3)
			codec.PutI32(data, 0, 1)                                // an item: all metadata is defined
			codec.PutI32(data, l.Stride, 2)                         // a list: bytes contain IDs
			codec.PutI32(data, l.Stride*2, int32(l.LeafCapacity+1)) // a split
			want := bytes.Clone(data)
			node := want[2*l.Stride:]
			if kind == codec.Dot {
				clear(node[16:24])
			} else {
				clear(node[12:16])
				clear(node[24:])
			}
			canonicalPadding(data, l)
			if !bytes.Equal(data, want) {
				t.Fatal("undefined bytes not cleared, or defined fields changed")
			}
			canonicalPadding(data, l)
			if !bytes.Equal(data, want) {
				t.Fatal("not idempotent")
			}
		})
	}
}
