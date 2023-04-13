package angular

import (
	"math"
	"unsafe"

	"github.com/mariotoffia/goannoy/distance"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/vector"
)

type angularDistanceImpl[TV interfaces.VectorType, TIX interfaces.IndexTypes] struct {
	nodeSize       TIX
	maxNumChildren TIX
	vectorLength   TIX
}

// Distance creates a new angular distance implementation.
func Distance[TV interfaces.VectorType, TIX interfaces.IndexTypes](
	vectorLength TIX,
) *angularDistanceImpl[TV, TIX] {

	n := AngularNodeImpl[TV, TIX]{}

	ad := &angularDistanceImpl[TV, TIX]{
		vectorLength: vectorLength,
		nodeSize: TIX(
			unsafe.Offsetof(n.v) +
				(uintptr(vectorLength) * unsafe.Sizeof(TV(0))),
		),
	}

	// _K = (S) (((size_t) (_s - offsetof(Node, children))) / sizeof(S));
	size := uintptr(ad.nodeSize) - unsafe.Offsetof(n.children)
	ad.maxNumChildren = TIX(size / unsafe.Sizeof(n.children[0]))

	return ad
}

func (a *angularDistanceImpl[TV, TIX]) VectorLength() TIX {
	return a.vectorLength
}

func (a *angularDistanceImpl[TV, TIX]) MaxNumChildren() TIX {
	return a.maxNumChildren
}

func (a *angularDistanceImpl[TV, TIX]) NodeSize() TIX {
	return a.nodeSize
}

func (a *angularDistanceImpl[TV, TIX]) MapNodeToMemory(
	mem unsafe.Pointer,
	itemIndex TIX,
) interfaces.Node[TV, TIX] {
	pos := unsafe.Add(mem, itemIndex*a.nodeSize)

	return (*AngularNodeImpl[TV, TIX])(pos)
}

func (a *angularDistanceImpl[TV, TIX]) PreProcess(nodes unsafe.Pointer, node_count TIX) {
	// DO NOTHING
}

func (a *angularDistanceImpl[TV, TIX]) Normalize(node interfaces.Node[TV, TIX]) {
	raw := node.GetRawVector()
	norm := TV(vector.GetNormUnsafe(raw, a.vectorLength))

	if norm > 0 {
		ptr := unsafe.Pointer(raw)
		size := TIX(unsafe.Sizeof(TV(0)))

		for i := TIX(0); i < a.vectorLength; i++ {
			f := (*TV)(unsafe.Pointer(unsafe.Add(ptr, i*size)))
			*f /= norm
		}
	}
}

func (a *angularDistanceImpl[TV, TIX]) Distance(x interfaces.Node[TV, TIX], y interfaces.Node[TV, TIX]) TV {
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

func (a *angularDistanceImpl[TV, TIX]) Margin(n interfaces.Node[TV, TIX], y []TV) TV {
	if len(y) == 0 {
		panic("y is empty")
	}

	return vector.DotUnsafe(
		n.GetRawVector(),
		(*TV)(unsafe.Pointer(unsafe.SliceData(y))),
		a.vectorLength,
	)
}

func (a *angularDistanceImpl[TV, TIX]) Side(
	n interfaces.Node[TV, TIX],
	y []TV,
	random interfaces.Random[TIX],
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

func (a *angularDistanceImpl[TV, TIX]) CreateSplit(
	nodes []interfaces.Node[TV, TIX],
	nodeSize TIX,
	random interfaces.Random[TIX],
	n interfaces.Node[TV, TIX],
) {
	// Allocate memory for two nodes, and use them as temporary nodes
	p_mem := make([]byte, nodeSize)
	q_mem := make([]byte, nodeSize)

	p := (*AngularNodeImpl[TV, TIX])(unsafe.Pointer(unsafe.SliceData(p_mem)))
	q := (*AngularNodeImpl[TV, TIX])(unsafe.Pointer(unsafe.SliceData(q_mem)))

	distance.TwoMeans[TV, TIX](nodes, a.vectorLength, random, true, p, q, a)

	nv := n.GetVector(a.vectorLength)
	qv := q.GetVector(a.vectorLength)
	pv := p.GetVector(a.vectorLength)

	for z := TIX(0); z < a.vectorLength; z++ {
		nv[z] = pv[z] - qv[z]
	}

	a.Normalize(n)
}

func (a *angularDistanceImpl[TV, TIX]) PQDistance(distance, margin TV, side interfaces.Side) TV {
	if side == interfaces.SideLeft {
		margin = -margin
	}
	return TV(math.Min(float64(distance), float64(margin)))
}

func (a *angularDistanceImpl[TV, TIX]) PQInitialValue() TV {
	return TV(math.Inf(1))
}

// InitNode will initialize the node by setting the norm to the value based on the distance type.
func (a *angularDistanceImpl[TV, TIX]) InitNode(node interfaces.Node[TV, TIX]) {
	norm := vector.DotUnsafe(node.GetRawVector(), node.GetRawVector(), a.vectorLength)
	node.SetNorm(norm)
}

func (a *angularDistanceImpl[TV, TIX]) Name() string {
	return "angular"
}
