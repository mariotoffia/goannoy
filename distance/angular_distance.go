package distance

import (
	"math"

	"github.com/mariotoffia/goannoy/random"
	"github.com/mariotoffia/goannoy/vector"
)

type AngularDistance[TV VectorType, TR random.RandomTypes] struct{}

func (a *AngularDistance[TV, TR]) NewNode(vectorLength int) *AngularNodeImpl[TV] {
	return &AngularNodeImpl[TV]{}
}

// PreProcess implements the `DistancePreprocessor` interface.
func (a *AngularDistance[TV, TR]) PreProcess(nodes []Node[TV], vectorLength int) {
	// DO NOTHING
}

func (a *AngularDistance[TV, TR]) Margin(n *AngularNodeImpl[TV], y [vector.ANNOYLIB_V_ARRAY_SIZE]TV, vectorLength int) TV {
	return vector.Dot(n.v, y, vectorLength)
}

// Side determines which side x or y.
func (a *AngularDistance[TV, TR]) Side(
	x *AngularNodeImpl[TV],
	y [vector.ANNOYLIB_V_ARRAY_SIZE]TV,
	random random.Random[TR],
	vectorLength int) bool {
	dotProduct := a.Margin(x, y, vectorLength)

	if dotProduct != 0 {
		return dotProduct > 0
	}

	return random.NextBool()
}

func (a *AngularDistance[TV, TR]) CreateSplit(
	nodes []Node[TV], vectorLength,
	s int,
	random random.Random[TR],
	n *AngularNodeImpl[TV]) {
	//
	p := a.NewNode(vectorLength)
	q := a.NewNode(vectorLength)

	twoMeans[TV](nodes, vectorLength, random, true, p, q)

	for z := 0; z < vectorLength; z++ {
		n.v[z] = p.v[z] - q.v[z]
	}

	n.Normalize(vectorLength)
}

func (a *AngularDistance[TV, TR]) NormalizedDistance(distance TV) TV {
	return TV(math.Sqrt(math.Max(float64(distance), 0)))
}

func (a *AngularDistance[TV, TR]) PQDistance(distance, margin TV, childNr int) TV {
	if childNr == 0 {
		margin = -margin
	}
	return TV(math.Min(float64(distance), float64(margin)))
}

func (a *AngularDistance[TV, TR]) PQInitialValue() TV {
	return math.MaxFloat32
}

func (a *AngularDistance[TV, TR]) Name() string {
	return "angular"
}

// AngularNodeImpl is a `NodeImpl` to be used with `AngularDistance`
type AngularNodeImpl[TV VectorType] struct {
	NodeImpl[TV]
	norm TV
}

// InitNode will initialize the node by setting the norm to
// the value based on the distance type.
func (n *AngularNodeImpl[TV]) InitNode(vectorLength int) {
	n.norm = vector.Dot(n.v, n.v, vectorLength)
}

func (n *AngularNodeImpl[TV]) CopyNodeTo(dst Node[TV], vectorLength int) {
	d := dst.(*AngularNodeImpl[TV])

	for z := 0; z < vectorLength; z++ {
		d.v[z] = n.v[z]
	}

	d.norm = n.norm
}

func (x *AngularNodeImpl[TV]) Distance(to Node[TV], vectorLength int) TV {
	t := to.(*AngularNodeImpl[TV])
	pp := x.norm

	if pp == 0 {
		pp = vector.Dot(x.v, x.v, vectorLength)
	}

	yv := t.v
	qq := t.norm

	if qq == 0 {
		qq = vector.Dot(yv, yv, vectorLength)
	}

	pq := vector.Dot(x.v, yv, vectorLength)
	ppqq := pp * qq

	if ppqq > 0 {
		return 2.0 - 2.0*pq/TV(math.Sqrt(float64(ppqq)))
	}
	return 2.0

}
