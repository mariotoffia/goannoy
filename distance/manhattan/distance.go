package manhattan

import (
	"math"
	"unsafe"

	"github.com/mariotoffia/goannoy/distance"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/vector"
)

type minkowskiNodeImpl[TV interfaces.VectorType] struct {
	n_descendants interfaces.ItemID
	a             TV
	children      [2]interfaces.ItemID
	v             [0]TV
}

func (n *minkowskiNodeImpl[TV]) GetRawVector() *TV {
	return (*TV)(unsafe.Pointer(&n.v))
}

func (n *minkowskiNodeImpl[TV]) GetVector(vectorLength int) []TV {
	return unsafe.Slice((*TV)(unsafe.Pointer(&n.v)), vectorLength)
}

func (n *minkowskiNodeImpl[TV]) SetVector(v []TV) {
	dst := unsafe.Pointer(&n.v)
	src := unsafe.Pointer(unsafe.SliceData(v))
	size := uintptr(len(v)) * unsafe.Sizeof(TV(0))
	copy((*[1 << 30]byte)(dst)[:size], (*[1 << 30]byte)(src)[:size])
}

func (n *minkowskiNodeImpl[TV]) GetRawChildren() *interfaces.ItemID {
	return (*interfaces.ItemID)(unsafe.Pointer(&n.children))
}

func (n *minkowskiNodeImpl[TV]) GetChildren() []interfaces.ItemID {
	if n.n_descendants == 0 {
		return nil
	}

	return unsafe.Slice((*interfaces.ItemID)(unsafe.Pointer(&n.children)), n.n_descendants)
}

func (n *minkowskiNodeImpl[TV]) SetChildren(children []interfaces.ItemID) {
	dst := unsafe.Pointer(&n.children)
	src := unsafe.Pointer(unsafe.SliceData(children))
	size := uintptr(len(children)) * unsafe.Sizeof(n.children[0])
	copy((*[1 << 30]byte)(dst)[:size], (*[1 << 30]byte)(src)[:size])
}

func (n *minkowskiNodeImpl[TV]) GetNumberOfDescendants() interfaces.ItemID {
	return n.n_descendants
}

func (n *minkowskiNodeImpl[TV]) SetNumberOfDescendants(nDescendants interfaces.ItemID) {
	n.n_descendants = nDescendants
}

func (n *minkowskiNodeImpl[TV]) GetNorm() TV {
	return n.a
}

func (n *minkowskiNodeImpl[TV]) SetNorm(norm TV) {
	n.a = norm
}

type manhattanDistanceImpl[TV interfaces.VectorType] struct {
	nodeSize       int
	maxNumChildren interfaces.ItemID
	vectorLength   int
}

func Distance[TV interfaces.VectorType](vectorLength int) *manhattanDistanceImpl[TV] {
	n := minkowskiNodeImpl[TV]{}
	d := &manhattanDistanceImpl[TV]{
		vectorLength: vectorLength,
		nodeSize:     int(unsafe.Offsetof(n.v) + uintptr(vectorLength)*unsafe.Sizeof(TV(0))),
	}

	size := uintptr(d.nodeSize) - unsafe.Offsetof(n.children)
	d.maxNumChildren = interfaces.ItemID(size / unsafe.Sizeof(n.children[0]))

	return d
}

func (d *manhattanDistanceImpl[TV]) PreProcess(nodes unsafe.Pointer, nodeCount interfaces.ItemID) {}

func (d *manhattanDistanceImpl[TV]) Distance(x interfaces.Node[TV], y interfaces.Node[TV]) TV {
	return vector.ManhattanDistance(x.GetVector(d.vectorLength), y.GetVector(d.vectorLength), d.vectorLength)
}

func (d *manhattanDistanceImpl[TV]) Normalize(node interfaces.Node[TV]) {
	raw := node.GetRawVector()
	norm := vector.GetNormUnsafe(raw, d.vectorLength)
	if !(norm > 0) {
		return
	}

	ptr := unsafe.Pointer(raw)
	size := unsafe.Sizeof(TV(0))
	for i := 0; i < d.vectorLength; i++ {
		f := (*TV)(unsafe.Pointer(unsafe.Add(ptr, uintptr(i)*size)))
		*f /= norm
	}
}

func (d *manhattanDistanceImpl[TV]) Margin(n interfaces.Node[TV], y []TV) TV {
	node := n.(*minkowskiNodeImpl[TV])
	return node.a + vector.DotUnsafe(node.GetRawVector(), (*TV)(unsafe.Pointer(unsafe.SliceData(y))), d.vectorLength)
}

func (d *manhattanDistanceImpl[TV]) CreateSplit(nodes []interfaces.Node[TV], nodeSize int, random interfaces.Random, n interfaces.Node[TV]) {
	pMem := make([]byte, nodeSize)
	qMem := make([]byte, nodeSize)
	p := (*minkowskiNodeImpl[TV])(unsafe.Pointer(unsafe.SliceData(pMem)))
	q := (*minkowskiNodeImpl[TV])(unsafe.Pointer(unsafe.SliceData(qMem)))

	distance.TwoMeans[TV](
		nodes,
		d.vectorLength,
		random,
		false,
		interfaces.Node[TV](p),
		interfaces.Node[TV](q),
		d,
	)

	nv := n.GetVector(d.vectorLength)
	pv := p.GetVector(d.vectorLength)
	qv := q.GetVector(d.vectorLength)
	for z := 0; z < d.vectorLength; z++ {
		nv[z] = pv[z] - qv[z]
	}

	d.Normalize(n)
	node := n.(*minkowskiNodeImpl[TV])
	node.a = 0
	for z := 0; z < d.vectorLength; z++ {
		node.a += -nv[z] * (pv[z] + qv[z]) / 2
	}
}

func (d *manhattanDistanceImpl[TV]) Side(n interfaces.Node[TV], y []TV, random interfaces.Random) interfaces.Side {
	margin := d.Margin(n, y)
	if margin != 0 {
		if margin > 0 {
			return interfaces.SideRight
		}
		return interfaces.SideLeft
	}
	return random.NextSide()
}

func (d *manhattanDistanceImpl[TV]) MapNodeToMemory(mem unsafe.Pointer, itemIndex interfaces.ItemID) interfaces.Node[TV] {
	pos := unsafe.Add(mem, uintptr(itemIndex)*uintptr(d.nodeSize))
	return (*minkowskiNodeImpl[TV])(pos)
}

func (d *manhattanDistanceImpl[TV]) PQDistance(distance, margin TV, side interfaces.Side) TV {
	if side == interfaces.SideLeft {
		margin = -margin
	}
	return TV(math.Min(float64(distance), float64(margin)))
}

func (d *manhattanDistanceImpl[TV]) NormalizedDistance(distance TV) TV {
	return TV(math.Max(float64(distance), 0))
}

func (d *manhattanDistanceImpl[TV]) PQInitialValue() TV {
	return TV(math.Inf(1))
}

func (d *manhattanDistanceImpl[TV]) InitNode(node interfaces.Node[TV]) {}

func (d *manhattanDistanceImpl[TV]) MaxNumChildren() interfaces.ItemID {
	return d.maxNumChildren
}

func (d *manhattanDistanceImpl[TV]) NodeSize() int {
	return d.nodeSize
}

func (d *manhattanDistanceImpl[TV]) VectorLength() int {
	return d.vectorLength
}
