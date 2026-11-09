package core

type Candidate struct {
	Priority float32
	ID       int32
}
type CandidateHeap []Candidate

func higher(a, b Candidate) bool {
	return a.Priority > b.Priority || a.Priority == b.Priority && a.ID > b.ID
}
func (h *CandidateHeap) Push(x Candidate) {
	*h = append(*h, x)
	n := len(*h) - 1
	for n > 0 {
		p := (n - 1) / 2
		if !higher(x, (*h)[p]) {
			break
		}
		(*h)[n] = (*h)[p]
		n = p
	}
	(*h)[n] = x
}
func (h *CandidateHeap) Pop() Candidate {
	a := *h
	top := a[0]
	last := a[len(a)-1]
	a = a[:len(a)-1]
	if len(a) > 0 {
		n := 0
		for c := 1; c < len(a); c = n*2 + 1 {
			if c+1 < len(a) && higher(a[c+1], a[c]) {
				c++
			}
			if !higher(a[c], last) {
				break
			}
			a[n] = a[c]
			n = c
		}
		a[n] = last
	}
	*h = a
	return top
}
