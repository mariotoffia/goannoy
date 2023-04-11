package angular

import (
	"math"
	"unsafe"

	"github.com/mariotoffia/goannoy/distance"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/vector"
)

type AngularDistanceImpl[TV interfaces.VectorType, TR interfaces.RandomTypes] struct {
	nodeSize       int
	maxNumChildren int
	vectorLength   int
}

func NewDistance[TV interfaces.VectorType, TR interfaces.RandomTypes](
	vectorType TV,
	randomType TR,
	vectorLength int,
) *AngularDistanceImpl[TV, TR] {

	n := AngularNodeImpl[TV]{}

	ad := &AngularDistanceImpl[TV, TR]{
		vectorLength: vectorLength,
		nodeSize: int(
			unsafe.Offsetof(n.v) +
				(uintptr(vectorLength) * unsafe.Sizeof(TV(0))),
		),
	}

	// _K = (S) (((size_t) (_s - offsetof(Node, children))) / sizeof(S));
	ad.maxNumChildren = int((uintptr(ad.nodeSize) -
		(unsafe.Offsetof(n.children))/unsafe.Sizeof(n.children[0])),
	)

	return ad
}

func (a *AngularDistanceImpl[TV, TR]) VectorLength() int {
	return a.vectorLength
}

func (a *AngularDistanceImpl[TV, TR]) MaxNumChildren() int {
	return a.maxNumChildren
}

func (a *AngularDistanceImpl[TV, TR]) NodeSize() int {
	return a.nodeSize
}

func (a *AngularDistanceImpl[TV, TR]) MapNodeToMemory(
	mem unsafe.Pointer,
	itemIndex int,
) interfaces.Node[TV] {
	pos := unsafe.Add(mem, itemIndex*a.nodeSize)

	return (*AngularNodeImpl[TV])(pos)
}

func (a *AngularDistanceImpl[TV, TR]) PreProcess(nodes unsafe.Pointer, node_count int) {
	// DO NOTHING
}

func (a *AngularDistanceImpl[TV, TR]) Normalize(node interfaces.Node[TV]) {
	raw := node.GetRawVector()
	norm := TV(vector.GetNormUnsafe(raw, a.vectorLength))

	if norm > 0 {
		ptr := unsafe.Pointer(raw)
		size := int(unsafe.Sizeof(TV(0)))

		for i := 0; i < a.vectorLength; i++ {
			f := (*TV)(unsafe.Pointer(unsafe.Add(ptr, i*size)))
			*f /= norm
		}
	}
}

func (a *AngularDistanceImpl[TV, TR]) Distance(x interfaces.Node[TV], y interfaces.Node[TV]) TV {
	pp := x.GetNorm()
	qq := y.GetNorm()

	if pp == 0 {
		pp = vector.DotUnsafe(x.GetRawVector(), x.GetRawVector(), a.vectorLength)
	}

	if qq == 0 {
		qq = vector.DotUnsafe(y.GetRawVector(), y.GetRawVector(), a.vectorLength)
	}

	pq := vector.DotUnsafe(x.GetRawVector(), y.GetRawVector(), a.vectorLength)
	ppqq := pp * qq

	if ppqq > 0 {
		return 2.0 - 2.0*pq/TV(math.Sqrt(float64(ppqq)))
	}
	return 2.0
}

func (a *AngularDistanceImpl[TV, TR]) Margin(n interfaces.Node[TV], y []TV) TV {
	if len(y) == 0 {
		panic("y is empty")
	}

	return vector.DotUnsafe(
		n.GetRawVector(),
		(*TV)(unsafe.Pointer(unsafe.SliceData(y))),
		a.vectorLength,
	)
}

func (a *AngularDistanceImpl[TV, TR]) Side(
	n interfaces.Node[TV],
	y []TV,
	random interfaces.Random[TR],
) interfaces.Side {

	dot := a.Margin(n, y)

	if dot != 0 {
		if dot > 0 {
			return interfaces.SideRight
		} else {
			return interfaces.SideLeft
		}
	}

	return random.NextSide()
}

func (a *AngularDistanceImpl[TV, TR]) CreateSplit(
	nodes []interfaces.Node[TV],
	nodeSize int,
	random interfaces.Random[TR],
	n interfaces.Node[TV],
) {
	// Allocate memory for two nodes, and use them as temporary nodes
	p_mem := make([]byte, nodeSize)
	q_mem := make([]byte, nodeSize)

	p := (*AngularNodeImpl[TV])(unsafe.Pointer(unsafe.SliceData(p_mem)))
	q := (*AngularNodeImpl[TV])(unsafe.Pointer(unsafe.SliceData(q_mem)))

	distance.TwoMeans[TV, TR](
		nodes,
		a.vectorLength,
		random,
		true,
		p, q,
		a,
	)

	nv := n.GetVector(a.vectorLength)
	qv := q.GetVector(a.vectorLength)
	pv := p.GetVector(a.vectorLength)

	for z := 0; z < a.vectorLength; z++ {
		nv[z] = pv[z] - qv[z]
	}

	a.Normalize(n)
}

func (a *AngularDistanceImpl[TV, TR]) PQDistance(distance, margin TV, side interfaces.Side) TV {
	if side == interfaces.SideLeft {
		margin = -margin
	}
	return TV(math.Min(float64(distance), float64(margin)))
}

func (a *AngularDistanceImpl[TV, TR]) PQInitialValue() TV {
	return TV(math.Inf(1))
}

// InitNode will initialize the node by setting the norm to the value based on the distance type.
func (a *AngularDistanceImpl[TV, TR]) InitNode(node interfaces.Node[TV]) {
	norm := vector.DotUnsafe(node.GetRawVector(), node.GetRawVector(), a.vectorLength)
	node.SetNorm(norm)
}

func (a *AngularDistanceImpl[TV, TR]) Name() string {
	return "angular"
}
