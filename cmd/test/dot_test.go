package test

import (
	"fmt"
	"testing"
)

func TestDotProduct(t *testing.T) {
	a := []float64{1, 2, 3}
	b := []float64{4, 5, 6}

	result := dot(&a[0], &b[0], len(a))
	fmt.Println("Dot product:", result)
}
