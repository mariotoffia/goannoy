package random

import (
	"math/rand"
	"time"
)

type GoRandom struct {
	rng  *rand.Rand
	seed uint32
}

func NewGoRandom() *GoRandom {
	return NewGoRandomWithSeed(uint32(time.Now().UnixNano()))
}

func NewGoRandomWithSeed(seed uint32) *GoRandom {
	src := rand.NewSource(time.Now().UnixNano())
	return &GoRandom{
		rng:  rand.New(src),
		seed: seed,
	}
}

func (r *GoRandom) Next() uint32 {
	return uint32(r.rng.Intn(0x7fffffff))
}

func (r *GoRandom) NextBool() bool {
	return r.rng.Intn(2) == 1
}

func (r *GoRandom) NormFloat64() float64 {
	return r.rng.NormFloat64()
}

func (r *GoRandom) NextIndex(n uint32) uint32 {
	return uint32(r.rng.Intn(int(n)))
}

func (r *GoRandom) GetSeed() uint32 {
	return r.seed
}

func (r *GoRandom) SetSeed(seed uint32) {
	r.rng.Seed(int64(seed))
}

func (r *GoRandom) CloneAndReset() *GoRandom {
	return NewGoRandomWithSeed(r.seed)
}
