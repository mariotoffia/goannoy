package core

import goannoy "github.com/mariotoffia/goannoy"

func Worse(a, b goannoy.Result) bool {
	return a.Distance > b.Distance || a.Distance == b.Distance && a.ID > b.ID
}

// Offer maintains a max heap in caller storage. The root is the worst result.
func Offer(dst []goannoy.Result, n int, r goannoy.Result) int {
	if n < len(dst) {
		i := n
		n++
		for i > 0 {
			p := (i - 1) / 2
			if !Worse(r, dst[p]) {
				break
			}
			dst[i] = dst[p]
			i = p
		}
		dst[i] = r
		return n
	}
	if n == 0 || !Worse(dst[0], r) {
		return n
	}
	i := 0
	for c := 1; c < n; c = i*2 + 1 {
		if c+1 < n && Worse(dst[c+1], dst[c]) {
			c++
		}
		if !Worse(dst[c], r) {
			break
		}
		dst[i] = dst[c]
		i = c
	}
	dst[i] = r
	return n
}
