package codec

import "testing"

func TestCheckFloatsAllocs(t *testing.T) {
	l, err := New(Angular, 3)
	if err != nil {
		t.Fatal(err)
	}
	v := []float32{1, 2, 3}
	if n := testing.AllocsPerRun(100, func() {
		if err := l.CheckFloats(v); err != nil {
			panic(err)
		}
	}); n != 0 {
		t.Fatalf("%g boundary-validation allocations", n)
	}
}

var floatSink float32

func TestFloatCodecAllocs(t *testing.T) {
	b := make([]byte, 8)
	PutF32(b, 0, 1.25)
	n := testing.AllocsPerRun(100, func() { floatSink = F32(b, 0); PutF32(b, 4, floatSink) })
	if n != 0 {
		t.Fatalf("float field access allocates %g times", n)
	}
	if F32(b, 4) != 1.25 {
		t.Fatal("float copy changed value")
	}
}
