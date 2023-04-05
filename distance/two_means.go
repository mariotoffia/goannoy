package distance

import (
	"github.com/mariotoffia/goannoy/random"
	"github.com/mariotoffia/goannoy/vector"
)

// TwoMeans is a helper function
//
// Note from the author:
//
// This algorithm is a huge heuristic. Empirically it works really well, but I
// can't motivate it well. The basic idea is to keep two centroids and assign
// points to either one of them. We weight each centroid by the number of points
// assigned to it, so to balance it.
func TwoMeans[TV VectorType, TR random.RandomTypes](
	nodes []Node[TV],
	vectorLength int,
	random random.Random[TR],
	cosine bool,
	p, q Node[TV]) {
	//
	const iterationSteps = 200

	nodeCount := uint32(len(nodes))

	i := random.NextIndex(TR(nodeCount))
	j := random.NextIndex(TR(nodeCount - 1))

	if j >= i {
		j++ // ensure that i != j
	}

	nodes[i].CopyNodeTo(p, vectorLength)
	nodes[j].CopyNodeTo(q, vectorLength)

	if cosine {
		p.Normalize(vectorLength)
		q.Normalize(vectorLength)
	}

	p.InitNode(vectorLength)
	p.InitNode(vectorLength)

	pvec := p.GetVector(vectorLength)
	qvec := q.GetVector(vectorLength)

	ic, jc := float64(1), float64(1)
	for l := 0; l < iterationSteps; l++ {
		k := random.NextIndex(TR(nodeCount))

		di := ic * float64(p.Distance(nodes[k], vectorLength))
		dj := jc * float64(q.Distance(nodes[k], vectorLength))

		var norm TV

		vec := nodes[k].GetVector(vectorLength)
		if cosine {
			norm = vector.GetNorm(vec, vectorLength)

			if !(norm > 0) {
				continue
			}

		} else {
			norm = 1
		}

		if di < dj {
			for z := 0; z < vectorLength; z++ {
				pvec[z] = (pvec[z]*TV(ic) + vec[z]/norm) / TV(ic+1)
			}

			p.InitNode(vectorLength)
			ic++

		} else if dj < di {
			for z := 0; z < vectorLength; z++ {
				qvec[z] = (qvec[z]*TV(jc) + vec[z]/norm) / TV(jc+1)
			}

			q.InitNode(vectorLength)
			jc++
		}
	}
}
