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
type Kiss64Random struct {
	x, y, z, c, seed uint64
}

func NewKiss64Random(seed uint64) *Kiss64Random {
	if seed == 0 {
		seed = 1234567890987654321
	}

	return &Kiss64Random{
		x:    seed,
		y:    362436362436362436,
		z:    1066149217761810,
		c:    123456123456123456,
		seed: seed,
	}
}

func (r *Kiss64Random) Next() uint64 {
	r.z = 6906969069*r.z + 1234567
	r.y ^= r.y << 13
	r.y ^= r.y >> 17
	r.y ^= r.y << 43

	t := r.x<<58 + r.c

	r.c = r.x >> 6
	r.x += t

	if r.x < t {
		r.c++
	}

	return r.x + r.y + r.z
}

func (r *Kiss64Random) NextBool() bool {
	return r.Next()&1 == 1
}

func (r *Kiss64Random) NextSide() interfaces.Side {
	if r.NextBool() {
		return interfaces.SideLeft
	}
	return interfaces.SideRight
}

func (r *Kiss64Random) NextIndex(n uint64) uint64 {
	return r.Next() % n
}

func (r *Kiss64Random) SetSeed(seed uint64) {
	r.x = seed
	r.seed = seed
}

func (r *Kiss64Random) GetSeed() uint64 {
	return r.seed
}

func (r *Kiss64Random) CloneAndReset() *Kiss64Random {
	return NewKiss64Random(r.seed)
}
