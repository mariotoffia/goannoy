package core

import (
	"errors"
	goannoy "github.com/mariotoffia/goannoy"
	"math"
	"testing"
)

// Catches invalid resource settings reaching memory allocation or tree traversal.
func TestBuildConfigRejectsOverflowAndInvalidValues(t *testing.T) {
	for _, dim := range []int{0, -1, math.MaxInt} {
		if _, err := NewBuildConfig(dim); !errors.Is(err, goannoy.ErrInvalidDimension) {
			t.Errorf("dimension %d: %v", dim, err)
		}
	}
	setters := []func(*BuildConfig){
		func(c *BuildConfig) { c.Trees(0) }, func(c *BuildConfig) { c.Trees(-1) },
		func(c *BuildConfig) { c.Workers(-1) }, func(c *BuildConfig) { c.Capacity(0) },
		func(c *BuildConfig) { c.Capacity(math.MaxInt) }, func(c *BuildConfig) { c.Profile(255) },
		func(c *BuildConfig) { c.OnDisk("") },
	}
	for n, set := range setters {
		c, err := NewBuildConfig(3)
		if err != nil {
			t.Fatal(err)
		}
		set(c)
		if !errors.Is(c.Err(), goannoy.ErrInvalidOption) {
			t.Errorf("case %d: %v", n, c.Err())
		}
		first := c.Err()
		c.Trees(3).Workers(0).Seed(42).Profile(goannoy.LowMemory).Capacity(64)
		if c.Err() != first {
			t.Fatal("configuration error was overwritten")
		}
	}
}

func TestConfigProfilePreservesOverrides(t *testing.T) {
	c, _ := NewBuildConfig(3)
	c.Workers(2).Capacity(100).Trees(3).Seed(42).Profile(goannoy.Throughput)
	r := c.Resolve()
	if r.Workers != 2 || r.Capacity != 100 || r.Trees != 3 || r.Seed != 42 {
		t.Fatalf("lost explicit overrides: %+v", r)
	}
	for _, profile := range []goannoy.Profile{goannoy.Balanced, goannoy.Throughput, goannoy.LowMemory} {
		c, _ := NewBuildConfig(3)
		c.Profile(profile)
		r := c.Resolve()
		if r.Workers < 1 || r.Trees < 1 {
			t.Fatalf("invalid defaults: %+v", r)
		}
	}
}
