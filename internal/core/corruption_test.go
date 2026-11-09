package core

import (
	"github.com/mariotoffia/goannoy/internal/codec"
	"math"
	"os"
	"testing"
)

func TestOpenRejectsCorruptMetricFields(t *testing.T) {
	for _, tc := range []struct {
		k    codec.Kind
		off  int
		bits uint32
	}{
		{codec.Angular, 4, math.Float32bits(-1)},
		{codec.Angular, 4, math.Float32bits(float32(math.NaN()))},
		{codec.Dot, 12, math.Float32bits(float32(math.Inf(1)))},
		{codec.Dot, 16, math.Float32bits(-1)},
	} {
		l, _ := codec.New(tc.k, 3)
		b, e := os.ReadFile("../../compat/fixtures/" + tc.k.String() + ".ann")
		if e != nil {
			t.Fatal(e)
		}
		codec.PutI32(b, tc.off, int32(tc.bits))
		if _, e := Validate(b, l); e == nil {
			t.Errorf("%s offset %d accepted corrupt field", tc.k, tc.off)
		}
	}
}
