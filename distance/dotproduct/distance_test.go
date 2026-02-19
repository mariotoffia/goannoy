package dotproduct

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	d := Distance[float32, uint32](3)
	assert.Equal(t, "dotproduct", d.Name())
}
