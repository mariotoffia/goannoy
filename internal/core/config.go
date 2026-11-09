package core

import (
	"fmt"
	goannoy "github.com/mariotoffia/goannoy"
	"math"
	"runtime"
)

// BuildConfig records explicit choices separately from profile defaults.
type BuildConfig struct {
	Dimension                int
	trees, workers, capacity int
	seed                     uint64
	profile                  goannoy.Profile
	path                     string
	workersSet               bool
	err                      error
}

// Settings is an immutable resolution of a builder configuration.
type Settings struct {
	Dimension, Trees, Workers, Capacity int
	Seed                                uint64
	Profile                             goannoy.Profile
	Path                                string
}

func NewBuildConfig(dimension int) (*BuildConfig, error) {
	if dimension < 1 || dimension > (math.MaxInt32-24)/8 {
		return nil, fmt.Errorf("%w: %d", goannoy.ErrInvalidDimension, dimension)
	}
	return &BuildConfig{Dimension: dimension, trees: 10, seed: 1234567890987654321}, nil
}
func (c *BuildConfig) fail(name string, value any) {
	if c.err == nil {
		c.err = fmt.Errorf("%w: %s=%v", goannoy.ErrInvalidOption, name, value)
	}
}
func (c *BuildConfig) Trees(n int) *BuildConfig {
	if n < 1 || n > math.MaxInt32 {
		c.fail("trees", n)
	} else {
		c.trees = n
	}
	return c
}
func (c *BuildConfig) Workers(n int) *BuildConfig {
	if n < 0 || n > math.MaxInt32 {
		c.fail("workers", n)
	} else {
		c.workers = n
		c.workersSet = true
	}
	return c
}
func (c *BuildConfig) Capacity(n int) *BuildConfig {
	if n < 1 || n > math.MaxInt32 || n > math.MaxInt/(24+8*c.Dimension) {
		c.fail("capacity", n)
	} else {
		c.capacity = n
	}
	return c
}
func (c *BuildConfig) Seed(seed uint64) *BuildConfig { c.seed = seed; return c }
func (c *BuildConfig) Profile(p goannoy.Profile) *BuildConfig {
	if p > goannoy.LowMemory {
		c.fail("profile", p)
	} else {
		c.profile = p
	}
	return c
}
func (c *BuildConfig) OnDisk(path string) *BuildConfig {
	if path == "" {
		c.fail("path", path)
	} else {
		c.path = path
	}
	return c
}
func (c *BuildConfig) Err() error { return c.err }
func (c *BuildConfig) Resolve() Settings {
	workers := runtime.GOMAXPROCS(0)
	if c.profile == goannoy.Balanced {
		workers = min(workers, 4)
	}
	if c.profile == goannoy.LowMemory {
		workers = 1
	}
	if c.workersSet {
		workers = c.workers
		if workers == 0 {
			workers = runtime.GOMAXPROCS(0)
		}
	}
	return Settings{Dimension: c.Dimension, Trees: c.trees, Workers: workers, Capacity: c.capacity, Seed: c.seed, Profile: c.profile, Path: c.path}
}
