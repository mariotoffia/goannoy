package core

import "testing"

func TestValueHeapOrderingAndGrowth(t *testing.T) {
	var h CandidateHeap
	for i := 0; i < 1000; i++ {
		h.Push(Candidate{Priority: float32(i % 7), ID: int32(i)})
	}
	prev := Candidate{Priority: 100, ID: 10000}
	for len(h) > 0 {
		got := h.Pop()
		if got.Priority > prev.Priority || got.Priority == prev.Priority && got.ID > prev.ID {
			t.Fatalf("%v after %v", got, prev)
		}
		prev = got
	}
}
