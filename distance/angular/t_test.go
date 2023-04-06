package angular

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type MyStruct struct {
	children [2]int
}

func (n *MyStruct) GetFloat64() float64 {
	return *(*float64)(unsafe.Pointer(&n.children))
}

func (n *MyStruct) SetFloat64(f float64) {
	*(*float64)(unsafe.Pointer(&n.children)) = f
}

func TestSetAndGetFloat64AndInt(t *testing.T) {
	n := &MyStruct{}
	n.SetFloat64(1.234)
	if n.GetFloat64() != 1.234 {
		t.Errorf("expected 1.234, got %v", n.GetFloat64())
	}

	n.children[0] = 1234
	n.children[1] = 5678

	assert.Equal(t, 1234, n.children[0])
	assert.Equal(t, 5678, n.children[1])
}

func TestArrValueWillBeOneFloat64(t *testing.T) {
	n := &MyStruct{children: [2]int{4607182418800017408, 0}}
	assert.Equal(t, float64(1), n.GetFloat64())
}
