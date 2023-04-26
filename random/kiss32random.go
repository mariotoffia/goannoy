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
type Kiss32Random[TIX interfaces.IndexTypes] struct {
	x, y, z, c, seed TIX
}

// NewKiss32Random creates a new random number generator based on the KISS
// algorithm.
func NewKiss32Random[TIX interfaces.IndexTypes](seed TIX) *Kiss32Random[TIX] {
	if seed == 0 {
		seed = 123456789
	}

	return &Kiss32Random[TIX]{
		x:    seed,
		y:    362436000,
		z:    521288629,
		c:    7654321,
		seed: seed,
	}
}

// Next returns the next random number.
func (r *Kiss32Random[TIX]) Next() TIX {
	r.x = 69069*r.x + 12345
	r.y ^= r.y << 13
	r.y ^= r.y >> 17
	r.y ^= r.y << 5

	t := uint64(698769069) + uint64(r.z) + uint64(r.c)

	r.c = TIX(t >> 32)
	r.z = TIX(t)

	return TIX(r.x + r.y + r.z)
}

func (r *Kiss32Random[TIX]) NextBool() bool {
	return r.Next()&1 == 1
}

func (r *Kiss32Random[TIX]) NextSide() interfaces.Side {
	if r.NextBool() {
		return interfaces.SideLeft
	}
	return interfaces.SideRight
}

func (r *Kiss32Random[TIX]) NextIndex(n TIX) TIX {
	return r.Next() % n
}

func (r *Kiss32Random[TIX]) SetSeed(seed TIX) {
	r.x = seed
	r.seed = seed
}

func (r *Kiss32Random[TIX]) GetSeed() TIX {
	return TIX(r.seed)
}

func (r *Kiss32Random[TIX]) CloneAndReset() interfaces.Random[TIX] {
	return NewKiss32Random(r.seed)
}
