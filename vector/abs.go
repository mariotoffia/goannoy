package vector

import (
	"github.com/mariotoffia/goannoy/interfaces"
)

// Abs is less performant on hardware where floating point operations are
// expensive, but it is generic.
func Abs[TV interfaces.VectorType](x TV) TV {
	if x < 0 {
		return -x
	}

	return x
}
