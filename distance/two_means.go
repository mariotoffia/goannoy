package distance

import (
	"github.com/mariotoffia/goannoy/amath"
	"github.com/mariotoffia/goannoy/random"
)

// twoMeans is a helper function
//
// Note from the author:
// This algorithm is a huge heuristic. Empirically it works really well, but I
// can't motivate it well. The basic idea is to keep two centroids and assign
// points to either one of them. We weight each centroid by the number of points
// assigned to it, so to balance it.
func twoMeans[TV VectorType, TR random.RandomTypes](
	nodes []Node[TV],
	vectorLength int,
	random random.Random[TR],
	cosine bool,
	p, q Node[TV],
	dist Distance[TV]) {
	const iterationSteps = 200

	nodeCount := uint32(len(nodes))

	i := random.NextIndex(TR(nodeCount))
	j := random.NextIndex(TR(nodeCount - 1))

	if j >= i {
		j++ // ensure that i != j
	}

	dist.CopyNode(p, nodes[i])
	dist.CopyNode(q, nodes[j])

	if cosine {
		dist.Normalize(p)
		dist.Normalize(q)
	}

	dist.InitNode(p)
	dist.InitNode(q)

	pvec := p.GetVector()
	qvec := q.GetVector()

	ic, jc := float64(1), float64(1)
	for l := 0; l < iterationSteps; l++ {
		k := random.NextIndex(TR(nodeCount))

		di := ic * float64(dist.Distance(p, nodes[k]))
		dj := jc * float64(dist.Distance(q, nodes[k]))

		var norm TV

		vec := nodes[k].GetVector()
		if cosine {
			norm = amath.GetNorm(vec)

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

			dist.InitNode(p)
			ic++

		} else if dj < di {
			for z := 0; z < vectorLength; z++ {
				qvec[z] = (qvec[z]*TV(jc) + vec[z]/norm) / TV(jc+1)
			}

			dist.InitNode(q)
			jc++
		}
	}
}
