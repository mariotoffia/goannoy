package main

import "github.com/mariotoffia/goannoy/internal/codec"

// canonicalPadding zeros only C++-uninitialized split fields, not algorithm data.
// It makes hashes reproducible without turning Go builds into the fixture oracle.
func canonicalPadding(data []byte, l codec.Layout) {
	for id := 0; id < len(data)/l.Stride; id++ {
		if int(l.Descendants(data, id)) <= l.LeafCapacity {
			continue
		}
		node := l.Node(data, id)
		switch l.Kind {
		case codec.Dot:
			clear(node[16:24])
		case codec.Hamming:
			clear(node[12:16])
			clear(node[24:])
		}
	}
}
