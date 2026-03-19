package distance

import (
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/utils"
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
func TwoMeans[TV interfaces.VectorType](
	nodes []interfaces.Node[TV],
	vectorLength int,
	random interfaces.Random,
	cosine bool,
	p, q interfaces.Node[TV],
	distance interfaces.Distance[TV],
) {

	const iterationSteps = 200

	nodeCount := interfaces.ItemID(len(nodes))

	i := random.NextIndex(nodeCount)
	j := random.NextIndex(nodeCount - 1)

	if j >= i {
		j++ // ensure that i != j
	}

	utils.CopyNode(p, nodes[i], distance.NodeSize())
	utils.CopyNode(q, nodes[j], distance.NodeSize())

	if cosine {
		distance.Normalize(p)
		distance.Normalize(q)
	}

	distance.InitNode(p)
	distance.InitNode(q)

	pvec := p.GetVector(vectorLength)
	qvec := q.GetVector(vectorLength)

	ic, jc := float64(1), float64(1)
	for l := 0; l < iterationSteps; l++ {
		k := random.NextIndex(nodeCount)

		di := ic * float64(distance.Distance(p, nodes[k]))
		dj := jc * float64(distance.Distance(q, nodes[k]))

		var norm TV

		vec := nodes[k].GetVector(vectorLength)

		if cosine {
			if nh, ok := distance.(interface {
				MeanNorm(node interfaces.Node[TV], vectorLength int) TV
			}); ok {
				norm = nh.MeanNorm(nodes[k], vectorLength)
			} else {
				norm = vector.GetNorm(vec, vectorLength)
			}

			if !(norm > 0) {
				continue
			}
		} else {
			norm = 1
		}

		if di < dj {
			if mh, ok := distance.(interface {
				UpdateMean(mean interfaces.Node[TV], newNode interfaces.Node[TV], norm TV, count int, vectorLength int)
			}); ok {
				mh.UpdateMean(p, nodes[k], norm, int(ic), vectorLength)
			} else {
				for z := 0; z < vectorLength; z++ {
					pvec[z] = (pvec[z]*TV(ic) + vec[z]/norm) / TV(ic+1)
				}
			}

			distance.InitNode(p)
			ic++

		} else if dj < di {
			if mh, ok := distance.(interface {
				UpdateMean(mean interfaces.Node[TV], newNode interfaces.Node[TV], norm TV, count int, vectorLength int)
			}); ok {
				mh.UpdateMean(q, nodes[k], norm, int(jc), vectorLength)
			} else {
				for z := 0; z < vectorLength; z++ {
					qvec[z] = (qvec[z]*TV(jc) + vec[z]/norm) / TV(jc+1)
				}
			}

			distance.InitNode(q)
			jc++
		}
	}
}
