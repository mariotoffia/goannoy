package random

import "github.com/mariotoffia/goannoy/interfaces"

// A random number generator based on the KISS algorithm.
//
// The KISS algorithm is a combination of four different random number
// generators. It is fast and has good randomness properties.
//
// The KISS algorithm is described in the paper "Good Parameter Sets for Combined
// Multiple Recursive Random Number Generators" by George Marsaglia.
//
// http://www0.cs.ucl.ac.uk/staff/d.jones/GoodPracticeRNG.pdf -> "Use a good RNG and build it into your code"
// http://mathforum.org/kb/message.jspa?messageID=6627731
// https://de.wikipedia.org/wiki/KISS_(Zufallszahlengenerator)
type Kiss32Random[T interfaces.RandomTypes] struct {
	x, y, z, c, seed T
}

// NewKiss32Random creates a new random number generator based on the KISS
// algorithm.
func NewKiss32Random[T interfaces.RandomTypes](seed T) *Kiss32Random[T] {
	if seed == 0 {
		seed = 123456789
	}

	return &Kiss32Random[T]{
		x:    seed,
		y:    362436000,
		z:    521288629,
		c:    7654321,
		seed: seed,
	}
}

// Next returns the next random number.
func (r *Kiss32Random[T]) Next() T {
	r.x = 69069*r.x + 12345
	r.y ^= r.y << 13
	r.y ^= r.y >> 17
	r.y ^= r.y << 5

	t := uint64(698769069) + uint64(r.z) + uint64(r.c)

	r.c = T(t >> 32)
	r.z = T(t)

	return T(r.x + r.y + r.z)
}

func (r *Kiss32Random[T]) NextBool() bool {
	return r.Next()&1 == 1
}

func (r *Kiss32Random[T]) NextSide() interfaces.Side {
	if r.NextBool() {
		return interfaces.SideLeft
	}
	return interfaces.SideRight
}

func (r *Kiss32Random[T]) NextIndex(n T) T {
	return r.Next() % n
}

func (r *Kiss32Random[T]) SetSeed(seed T) {
	r.x = seed
	r.seed = seed
}

func (r *Kiss32Random[T]) GetSeed() T {
	return T(r.seed)
}

func (r *Kiss32Random[T]) CloneAndReset() interfaces.Random[T] {
	return NewKiss32Random(r.seed)
}
