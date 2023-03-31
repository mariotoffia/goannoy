package amath

import (
	"math/rand"
	"time"
)

// rng is used in calc to generate random numbers
var rng *rand.Rand

func init() {
	// Create a new random number generator and seed it with the current time
	src := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(src)
}

func RandBool() bool {
	return rng.Intn(2) == 1
}
