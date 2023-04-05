package angular

import (
	"math"
	"unsafe"

	"github.com/mariotoffia/goannoy/distance"
	"github.com/mariotoffia/goannoy/random"
	"github.com/mariotoffia/goannoy/vector"
)

type AngularDistanceImpl[TV distance.VectorType, TR random.RandomTypes] struct{}

func (a *AngularDistanceImpl[TV, TR]) NewNode(vectorLength int) *AngularNodeImpl[TV] {
	return &AngularNodeImpl[TV]{}
}

// PreProcess implements the `interfaces.DistancePreprocessor` interface.
func (a *AngularDistanceImpl[TV, TR]) PreProcess(nodes []distance.Node[TV], vectorLength int) {
	// DO NOTHING
}

func (a *AngularDistanceImpl[TV, TR]) Margin(n *AngularNodeImpl[TV], y []TV, vectorLength int) TV {
	if len(y) == 0 {
		panic("y is empty")
	}

	return vector.DotUnsafe(n.GetRawVector(), (*TV)(unsafe.Pointer(&y[0])), vectorLength)
}

// Side determines which side x or y.
func (a *AngularDistanceImpl[TV, TR]) Side(
	x *AngularNodeImpl[TV],
	y []TV,
	random random.Random[TR],
	vectorLength int,
) bool {

	dotProduct := a.Margin(x, y, vectorLength)

	if dotProduct != 0 {
		return dotProduct > 0
	}

	return random.NextBool()
}

func (a *AngularDistanceImpl[TV, TR]) CreateSplit(
	nodes []distance.Node[TV], vectorLength,
	s int,
	random random.Random[TR],
	n *AngularNodeImpl[TV],
) {

	p := a.NewNode(vectorLength)
	q := a.NewNode(vectorLength)

	distance.TwoMeans[TV](nodes, vectorLength, random, true, p, q)

	for z := 0; z < vectorLength; z++ {
		n.v[z] = p.v[z] - q.v[z]
	}

	n.Normalize(vectorLength)
}

// NormalizeDistance implements the `interfaces.DistanceNormalizer` interface.
func (a *AngularDistanceImpl[TV, TR]) NormalizedDistance(distance TV) TV {
	return TV(math.Sqrt(math.Max(float64(distance), 0)))
}

func (a *AngularDistanceImpl[TV, TR]) PQDistance(distance, margin TV, childNr int) TV {
	if childNr == 0 {
		margin = -margin
	}
	return TV(math.Min(float64(distance), float64(margin)))
}

func (a *AngularDistanceImpl[TV, TR]) PQInitialValue() TV {
	return math.MaxFloat32
}

func (a *AngularDistanceImpl[TV, TR]) Name() string {
	return "angular"
}
