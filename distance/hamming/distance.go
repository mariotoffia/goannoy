package hamming

import (
	"math"
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
)

type hammingNodeImpl[TV interfaces.VectorType] struct {
	n_descendants interfaces.ItemID
	children      [2]interfaces.ItemID
	v             [0]TV
}

func (n *hammingNodeImpl[TV]) GetRawVector() *TV {
	return (*TV)(unsafe.Pointer(&n.v))
}

func (n *hammingNodeImpl[TV]) GetVector(vectorLength int) []TV {
	return unsafe.Slice((*TV)(unsafe.Pointer(&n.v)), vectorLength)
}

func (n *hammingNodeImpl[TV]) SetVector(v []TV) {
	dst := unsafe.Pointer(&n.v)
	src := unsafe.Pointer(unsafe.SliceData(v))
	size := uintptr(len(v)) * unsafe.Sizeof(TV(0))
	copy((*[1 << 30]byte)(dst)[:size], (*[1 << 30]byte)(src)[:size])
}

func (n *hammingNodeImpl[TV]) GetRawChildren() *interfaces.ItemID {
	return (*interfaces.ItemID)(unsafe.Pointer(&n.children))
}

func (n *hammingNodeImpl[TV]) GetChildren() []interfaces.ItemID {
	if n.n_descendants == 0 {
		return nil
	}

	return unsafe.Slice((*interfaces.ItemID)(unsafe.Pointer(&n.children)), n.n_descendants)
}

func (n *hammingNodeImpl[TV]) SetChildren(children []interfaces.ItemID) {
	dst := unsafe.Pointer(&n.children)
	src := unsafe.Pointer(unsafe.SliceData(children))
	size := uintptr(len(children)) * unsafe.Sizeof(n.children[0])
	copy((*[1 << 30]byte)(dst)[:size], (*[1 << 30]byte)(src)[:size])
}

func (n *hammingNodeImpl[TV]) GetNumberOfDescendants() interfaces.ItemID {
	return n.n_descendants
}

func (n *hammingNodeImpl[TV]) SetNumberOfDescendants(nDescendants interfaces.ItemID) {
	n.n_descendants = nDescendants
}

func (n *hammingNodeImpl[TV]) GetNorm() TV {
	return *(*TV)(unsafe.Pointer(&n.children))
}

func (n *hammingNodeImpl[TV]) SetNorm(norm TV) {
	*(*TV)(unsafe.Pointer(&n.children)) = norm
}

type hammingDistanceImpl[TV interfaces.VectorType] struct {
	nodeSize       int
	maxNumChildren interfaces.ItemID
	vectorLength   int
}

func Distance[TV interfaces.VectorType](vectorLength int) *hammingDistanceImpl[TV] {
	n := hammingNodeImpl[TV]{}
	d := &hammingDistanceImpl[TV]{
		vectorLength: vectorLength,
		nodeSize:     int(unsafe.Offsetof(n.v) + uintptr(vectorLength)*unsafe.Sizeof(TV(0))),
	}

	size := uintptr(d.nodeSize) - unsafe.Offsetof(n.children)
	d.maxNumChildren = interfaces.ItemID(size / unsafe.Sizeof(n.children[0]))

	return d
}

func (d *hammingDistanceImpl[TV]) PreProcess(nodes unsafe.Pointer, nodeCount interfaces.ItemID) {}

func bitValue[T interfaces.VectorType](v T) bool {
	return v > 0.5
}

func (d *hammingDistanceImpl[TV]) Distance(x interfaces.Node[TV], y interfaces.Node[TV]) TV {
	xv := x.GetVector(d.vectorLength)
	yv := y.GetVector(d.vectorLength)

	var dist TV
	for i := range xv {
		if bitValue(xv[i]) != bitValue(yv[i]) {
			dist++
		}
	}

	return dist
}

func (d *hammingDistanceImpl[TV]) Normalize(node interfaces.Node[TV]) {}

func (d *hammingDistanceImpl[TV]) Margin(n interfaces.Node[TV], y []TV) TV {
	bit := int(n.GetVector(d.vectorLength)[0])
	if bit >= 0 && bit < len(y) && bitValue(y[bit]) {
		return 1
	}
	return 0
}

func (d *hammingDistanceImpl[TV]) CreateSplit(nodes []interfaces.Node[TV], nodeSize int, random interfaces.Random, n interfaces.Node[TV]) {
	const maxIterations = 20

	curSize := 0
	i := 0
	dim := int(d.vectorLength)

	for ; i < maxIterations; i++ {
		n.GetVector(d.vectorLength)[0] = TV(random.NextIndex(interfaces.ItemID(dim)))
		curSize = 0
		for _, node := range nodes {
			if d.Margin(n, node.GetVector(d.vectorLength)) != 0 {
				curSize++
			}
		}
		if curSize > 0 && curSize < len(nodes) {
			return
		}
	}

	for j := 0; j < dim; j++ {
		n.GetVector(d.vectorLength)[0] = TV(j)
		curSize = 0
		for _, node := range nodes {
			if d.Margin(n, node.GetVector(d.vectorLength)) != 0 {
				curSize++
			}
		}
		if curSize > 0 && curSize < len(nodes) {
			return
		}
	}
}

func (d *hammingDistanceImpl[TV]) Side(n interfaces.Node[TV], y []TV, random interfaces.Random) interfaces.Side {
	if d.Margin(n, y) != 0 {
		return interfaces.SideRight
	}
	return interfaces.SideLeft
}

func (d *hammingDistanceImpl[TV]) MapNodeToMemory(mem unsafe.Pointer, itemIndex interfaces.ItemID) interfaces.Node[TV] {
	pos := unsafe.Add(mem, uintptr(itemIndex)*uintptr(d.nodeSize))
	return (*hammingNodeImpl[TV])(pos)
}

func (d *hammingDistanceImpl[TV]) PQDistance(distance, margin TV, side interfaces.Side) TV {
	if margin != TV(side) {
		return distance - 1
	}
	return distance
}

func (d *hammingDistanceImpl[TV]) NormalizedDistance(distance TV) TV {
	return distance
}

func (d *hammingDistanceImpl[TV]) PQInitialValue() TV {
	return TV(math.Inf(1))
}

func (d *hammingDistanceImpl[TV]) InitNode(node interfaces.Node[TV]) {}

func (d *hammingDistanceImpl[TV]) MaxNumChildren() interfaces.ItemID {
	return d.maxNumChildren
}

func (d *hammingDistanceImpl[TV]) NodeSize() int {
	return d.nodeSize
}

func (d *hammingDistanceImpl[TV]) VectorLength() int {
	return d.vectorLength
}
