package angular

import (
	"math"
	"unsafe"

	"github.com/mariotoffia/goannoy/distance"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/vector"
)

type angularDistanceImpl[TV interfaces.VectorType] struct {
	nodeSize       int
	maxNumChildren interfaces.ItemID
	vectorLength   int
}

// Distance creates a new angular distance implementation.
func Distance[TV interfaces.VectorType](vectorLength int) *angularDistanceImpl[TV] {

	n := AngularNodeImpl[TV]{}

	ad := &angularDistanceImpl[TV]{
		vectorLength: vectorLength,
		nodeSize: int(
			unsafe.Offsetof(n.v) +
				(uintptr(vectorLength) * unsafe.Sizeof(TV(0))),
		),
	}

	// _K = (S) (((size_t) (_s - offsetof(Node, children))) / sizeof(S));
	size := uintptr(ad.nodeSize) - unsafe.Offsetof(n.children)
	ad.maxNumChildren = interfaces.ItemID(size / unsafe.Sizeof(n.children[0]))

	return ad
}

func (a *angularDistanceImpl[TV]) VectorLength() int {
	return a.vectorLength
}

func (a *angularDistanceImpl[TV]) MaxNumChildren() interfaces.ItemID {
	return a.maxNumChildren
}

func (a *angularDistanceImpl[TV]) NodeSize() int {
	return a.nodeSize
}

func (a *angularDistanceImpl[TV]) MapNodeToMemory(
	mem unsafe.Pointer,
	itemIndex interfaces.ItemID,
) interfaces.Node[TV] {
	pos := unsafe.Add(mem, uintptr(itemIndex)*uintptr(a.nodeSize))

	return (*AngularNodeImpl[TV])(pos)
}

func (a *angularDistanceImpl[TV]) PreProcess(nodes unsafe.Pointer, nodeCount interfaces.ItemID) {
	// DO NOTHING
}

func (a *angularDistanceImpl[TV]) Normalize(node interfaces.Node[TV]) {
	raw := node.GetRawVector()
	norm := TV(vector.GetNormUnsafe(raw, a.vectorLength))

	if norm > 0 {
		ptr := unsafe.Pointer(raw)
		size := unsafe.Sizeof(TV(0))

		for i := 0; i < a.vectorLength; i++ {
			f := (*TV)(unsafe.Pointer(unsafe.Add(ptr, uintptr(i)*size)))
			*f /= norm
		}
	}
}

func (a *angularDistanceImpl[TV]) Distance(x interfaces.Node[TV], y interfaces.Node[TV]) TV {
	pp := x.GetNorm()
	qq := y.GetNorm()
	xv := x.GetRawVector()
	yv := y.GetRawVector()

	if pp == 0 {
		pp = vector.DotUnsafe(xv, xv, a.vectorLength)
	}

	if qq == 0 {
		qq = vector.DotUnsafe(yv, yv, a.vectorLength)
	}

	var ppqq TV

	if pp != 0 {
		ppqq = pp * qq
	}

	if ppqq > 0 {
		pq := vector.DotUnsafe(xv, yv, a.vectorLength)
		return 2.0 - 2.0*pq/TV(math.Sqrt(float64(ppqq)))
	}
	return 2.0
}

func (a *angularDistanceImpl[TV]) Margin(n interfaces.Node[TV], y []TV) TV {
	if len(y) == 0 {
		panic("y is empty")
	}

	return vector.DotUnsafe(
		n.GetRawVector(),
		(*TV)(unsafe.Pointer(unsafe.SliceData(y))),
		a.vectorLength,
	)
}

func (a *angularDistanceImpl[TV]) Side(
	n interfaces.Node[TV],
	y []TV,
	random interfaces.Random,
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

func (a *angularDistanceImpl[TV]) CreateSplit(
	nodes []interfaces.Node[TV],
	nodeSize int,
	random interfaces.Random,
	n interfaces.Node[TV],
) {
	// Allocate memory for two nodes, and use them as temporary nodes
	p_mem := make([]byte, nodeSize)
	q_mem := make([]byte, nodeSize)

	p := (*AngularNodeImpl[TV])(unsafe.Pointer(unsafe.SliceData(p_mem)))
	q := (*AngularNodeImpl[TV])(unsafe.Pointer(unsafe.SliceData(q_mem)))

	distance.TwoMeans[TV](nodes, a.vectorLength, random, true, p, q, a)

	nv := n.GetVector(a.vectorLength)
	qv := q.GetVector(a.vectorLength)
	pv := p.GetVector(a.vectorLength)

	for z := 0; z < a.vectorLength; z++ {
		nv[z] = pv[z] - qv[z]
	}

	a.Normalize(n)
}

func (a *angularDistanceImpl[TV]) NormalizedDistance(distance TV) TV {
	return TV(math.Sqrt(float64(distance)))
}

func (a *angularDistanceImpl[TV]) PQDistance(distance, margin TV, side interfaces.Side) TV {
	if side == interfaces.SideLeft {
		margin = -margin
	}
	return TV(math.Min(float64(distance), float64(margin)))
}

func (a *angularDistanceImpl[TV]) PQInitialValue() TV {
	return TV(math.Inf(1))
}

// InitNode will initialize the node by setting the norm to the value based on the distance type.
func (a *angularDistanceImpl[TV]) InitNode(node interfaces.Node[TV]) {
	norm := vector.DotUnsafe(node.GetRawVector(), node.GetRawVector(), a.vectorLength)
	node.SetNorm(norm)
}

func (a *angularDistanceImpl[TV]) Name() string {
	return "angular"
}
