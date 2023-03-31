package index

import (
	"github.com/mariotoffia/goannoy/amath"
	"github.com/mariotoffia/goannoy/random"
)

type AnnoyIndexImpl[
	T int,
	TVecType amath.Calculable,
	TRandType random.RandomTypes,
	TRand random.Random[TRandType]] struct {
}
