package core

import (
	"errors"
	"os"
	"testing"

	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
)

func TestRejectSharedSplitDAG(t *testing.T) {
	l, _ := codec.New(codec.Angular, 3)
	fixture, err := os.ReadFile("../../compat/fixtures/angular.ann")
	if err != nil {
		t.Fatal(err)
	}
	for _, shared := range []bool{false, true} {
		data := append([]byte(nil), fixture[:64*l.Stride]...)
		codec.PutI32(data, 0, 0) // a sparse hole never advances candidate budget
		previous, older := int32(0), int32(1)
		for depth := 0; depth < 20; depth++ {
			node := make([]byte, l.Stride)
			n := 6 + depth
			if depth == 19 {
				n = 64
			}
			codec.PutI32(node, 0, int32(n))
			codec.PutI32(node, 4, previous)
			other := previous
			if shared {
				other = older
			}
			codec.PutI32(node, 8, other)
			older, previous = previous, int32(len(data)/l.Stride)
			data = append(data, node...)
		}
		if _, err := Validate(data, l); !errors.Is(err, goannoy.ErrCorruptIndex) {
			t.Errorf("accepted exponential DAG (distinct edges=%v): %v", shared, err)
		}
	}
}
