package utils

import "time"

func Measure(f func()) (elapsed time.Duration) {
	start := time.Now()
	f()
	elapsed = time.Since(start)
	return
}

func MeasureWithReturn[T any](f func() T) (elapsed time.Duration, ret T) {
	start := time.Now()
	ret = f()
	elapsed = time.Since(start)
	return
}
