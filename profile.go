package goannoy

// Profile selects resource defaults. Explicit builder settings take precedence.
type Profile uint8

const (
	Balanced Profile = iota
	Throughput
	LowMemory
)

// OpenOption configures storage when opening an index.
type OpenOption uint8

// Prefault touches each mapped page before Open returns.
const Prefault OpenOption = 1
