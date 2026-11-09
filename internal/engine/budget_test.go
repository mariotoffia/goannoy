package engine

import (
	"errors"
	"testing"

	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
)

func TestSearchRejectsZeroBudget(t *testing.T) {
	for _, kind := range []codec.Kind{codec.Angular, codec.Dot, codec.Euclidean, codec.Manhattan, codec.Hamming} {
		t.Run(kind.String(), func(t *testing.T) {
			dim := 3
			if kind == codec.Hamming {
				dim = 130
			}
			i, err := Open("../../compat/fixtures/"+kind.String()+".ann", kind, dim)
			if err != nil {
				t.Fatal(err)
			}
			defer i.Close()
			s := i.NewSearcher()
			for _, count := range []int{0, 10} {
				dst := make([]goannoy.Result, count)
				if _, err := s.SearchID(0, dst, 0); !errors.Is(err, goannoy.ErrInvalidSearchK) {
					t.Errorf("ID count=%d: %v", count, err)
				}
				if kind == codec.Hamming {
					_, err = s.SearchWords([]uint64{1, 2, 1}, dst, 0)
				} else {
					_, err = s.SearchFloat([]float32{1, 2, 3}, dst, 0)
				}
				if !errors.Is(err, goannoy.ErrInvalidSearchK) {
					t.Errorf("vector count=%d: %v", count, err)
				}
			}
		})
	}
}
