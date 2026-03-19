package random

import (
	"math/rand"
	"time"

	"github.com/mariotoffia/goannoy/interfaces"
)

type GoRandom struct {
	rng  *rand.Rand
	seed uint64
}

func NewGoRandom() *GoRandom {
	return NewGoRandomWithSeed(uint64(time.Now().UnixNano()))
}

func NewGoRandomWithSeed(seed uint64) *GoRandom {
	src := rand.NewSource(int64(seed))
	return &GoRandom{
		rng:  rand.New(src),
		seed: seed,
	}
}

func (r *GoRandom) Next() uint64 {
	return uint64(r.rng.Int63())
}

func (r *GoRandom) NextBool() bool {
	return r.rng.Intn(2) == 1
}

func (r *GoRandom) NextSide() interfaces.Side {
	if r.NextBool() {
		return interfaces.SideLeft
	}
	return interfaces.SideRight
}

func (r *GoRandom) NormFloat64() float64 {
	return r.rng.NormFloat64()
}

func (r *GoRandom) NextIndex(n interfaces.ItemID) interfaces.ItemID {
	return interfaces.ItemID(r.rng.Intn(int(n)))
}

func (r *GoRandom) GetSeed() uint64 {
	return r.seed
}

func (r *GoRandom) SetSeed(seed uint64) {
	r.rng.Seed(int64(seed))
	r.seed = seed
}

func (r *GoRandom) CloneAndReset() interfaces.Random {
	return NewGoRandomWithSeed(r.seed)
}
