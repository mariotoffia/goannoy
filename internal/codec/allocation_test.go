package codec

import "testing"

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
