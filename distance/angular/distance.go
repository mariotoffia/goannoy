package angular

import (
	"math"
	"unsafe"

	"github.com/mariotoffia/goannoy/distance"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/vector"
)

type AngularDistanceImpl[TV interfaces.VectorType, TR interfaces.RandomTypes] struct{}

func NewDistance[TV interfaces.VectorType, TR interfaces.RandomTypes](
	vectorType TV,
	randomType TR,
) *AngularDistanceImpl[TV, TR] {
	return &AngularDistanceImpl[TV, TR]{}
}

func (a *AngularDistanceImpl[TV, TR]) MapNodeToMemory(
	mem unsafe.Pointer,
	itemIndex, vectorLength int,
) interfaces.Node[TV] {
	size := (*AngularNodeImpl[TV])(nil).Size(vectorLength)
	pos := unsafe.Add(mem, itemIndex*size)

	return (*AngularNodeImpl[TV])(pos)
}

func (a *AngularDistanceImpl[TV, TR]) NewNodeFromGC(vectorLength int) interfaces.Node[TV] {
	return &AngularNodeImpl[TV]{}
}

// PreProcess implements the `interfaces.DistancePreprocessor` interface.
func (a *AngularDistanceImpl[TV, TR]) PreProcess(nodes unsafe.Pointer, node_count, vectorLength int) {
	// DO NOTHING
}

func (a *AngularDistanceImpl[TV, TR]) Margin(n *AngularNodeImpl[TV], y []TV, vectorLength int) TV {
	if len(y) == 0 {
		panic("y is empty")
	}

	return vector.DotUnsafe(n.GetRawVector(), (*TV)(unsafe.Pointer(&y[0])), vectorLength)
}

// Side determines which side of the children indices to use when a split is made.
func (a *AngularDistanceImpl[TV, TR]) Side(
	m interfaces.Node[TV],
	v []TV,
	random interfaces.Random[TR],
	vectorLength int,
) interfaces.Side {

	dotProduct := a.Margin(m.(*AngularNodeImpl[TV]), v, vectorLength)

	if dotProduct != 0 {
		if dotProduct > 0 {
			return interfaces.SideRight
		} else {
			return interfaces.SideLeft
		}
	}

	return random.NextSide()
}

func (a *AngularDistanceImpl[TV, TR]) CreateSplit(
	children []interfaces.Node[TV],
	vectorLength, nodeSize int,
	random interfaces.Random[TR],
	m interfaces.Node[TV],
) {

	p_mem := make([]byte, nodeSize)
	q_mem := make([]byte, nodeSize)

	p := (*AngularNodeImpl[TV])(unsafe.Pointer(&p_mem[0]))
	q := (*AngularNodeImpl[TV])(unsafe.Pointer(&q_mem[0]))

	distance.TwoMeans[TV](children, vectorLength, random, true, p, q)

	mv := m.GetVector(vectorLength)

	for z := 0; z < vectorLength; z++ {
		mv[z] = p.v[z] - q.v[z]
	}

	m.Normalize(vectorLength)
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
