package dotproduct

import (
	"math"
	"unsafe"

	"github.com/mariotoffia/goannoy/distance"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/vector"
)

type dotProductDistanceImpl[TV interfaces.VectorType] struct {
	nodeSize       int
	maxNumChildren interfaces.ItemID
	vectorLength   int
}

// Distance creates a new dot product distance implementation.
func Distance[TV interfaces.VectorType](vectorLength int) *dotProductDistanceImpl[TV] {

	n := DotProductNodeImpl[TV]{}

	ad := &dotProductDistanceImpl[TV]{
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

func (dp *dotProductDistanceImpl[TV]) VectorLength() int {
	return dp.vectorLength
}

func (dp *dotProductDistanceImpl[TV]) MaxNumChildren() interfaces.ItemID {
	return dp.maxNumChildren
}

func (dp *dotProductDistanceImpl[TV]) NodeSize() int {
	return dp.nodeSize
}

func (dp *dotProductDistanceImpl[TV]) MapNodeToMemory(
	mem unsafe.Pointer,
	itemIndex interfaces.ItemID,
) interfaces.Node[TV] {
	pos := unsafe.Add(mem, uintptr(itemIndex)*uintptr(dp.nodeSize))

	return (*DotProductNodeImpl[TV])(pos)
}

func (dp *dotProductDistanceImpl[TV]) CreateSplit(
	nodes []interfaces.Node[TV],
	nodeSize int,
	random interfaces.Random,
	n interfaces.Node[TV],
) {
	// Allocate memory for two nodes, and use them as temporary nodes
	p_mem := make([]byte, nodeSize)
	q_mem := make([]byte, nodeSize)

	p := (*DotProductNodeImpl[TV])(unsafe.Pointer(unsafe.SliceData(p_mem)))
	q := (*DotProductNodeImpl[TV])(unsafe.Pointer(unsafe.SliceData(q_mem)))

	distance.TwoMeans[TV](nodes, dp.vectorLength, random, true, p, q, dp)

	nv := n.GetVector(dp.vectorLength)
	qv := q.GetVector(dp.vectorLength)
	pv := p.GetVector(dp.vectorLength)

	for z := 0; z < dp.vectorLength; z++ {
		nv[z] = pv[z] - qv[z]
	}

	dp.Normalize(n)
}

func (dp *dotProductDistanceImpl[TV]) Normalize(node interfaces.Node[TV]) {
	raw := node.GetRawVector()
	norm := TV(vector.GetNormUnsafe(raw, dp.vectorLength))

	if norm > 0 {
		ptr := unsafe.Pointer(raw)
		size := unsafe.Sizeof(TV(0))

		for i := 0; i < dp.vectorLength; i++ {
			f := (*TV)(unsafe.Pointer(unsafe.Add(ptr, uintptr(i)*size)))
			*f /= norm
		}

		node.(*DotProductNodeImpl[TV]).dot_factor /= norm
	}
}

func (dp *dotProductDistanceImpl[TV]) Margin(n interfaces.Node[TV], y []TV) TV {
	if len(y) == 0 {
		panic("y is empty")
	}

	return vector.DotUnsafe(
		n.GetRawVector(),
		(*TV)(unsafe.Pointer(unsafe.SliceData(y))),
		dp.vectorLength,
	)
}

func (dp *dotProductDistanceImpl[TV]) Side(
	n interfaces.Node[TV],
	y []TV,
	random interfaces.Random,
) interfaces.Side {

	dot := dp.Margin(n, y)

	if dot != 0 {
		if dot > 0 {
			return interfaces.SideRight
		} else {
			return interfaces.SideLeft
		}
	}

	return random.NextSide()
}

func (dp *dotProductDistanceImpl[TV]) NormalizedDistance(distance TV) TV {
	return -distance
}

func (dp *dotProductDistanceImpl[TV]) PQDistance(distance, margin TV, side interfaces.Side) TV {
	if side == interfaces.SideLeft {
		margin = -margin
	}
	return TV(math.Min(float64(distance), float64(margin)))
}

func (dp *dotProductDistanceImpl[TV]) PQInitialValue() TV {
	return TV(math.Inf(1))
}

func (dp *dotProductDistanceImpl[TV]) Distance(x interfaces.Node[TV], y interfaces.Node[TV]) TV {
	xn := x.(*DotProductNodeImpl[TV])
	yn := y.(*DotProductNodeImpl[TV])

	if xn.built || yn.built {
		return -vector.DotUnsafe(x.GetRawVector(), y.GetRawVector(), dp.vectorLength)
	}

	pp := x.GetNorm()
	qq := y.GetNorm()
	xv := x.GetRawVector()
	yv := y.GetRawVector()

	xdf := xn.dot_factor
	ydf := yn.dot_factor

	if pp == 0 {
		pp = vector.DotUnsafe(xv, xv, dp.vectorLength) + xdf*xdf
	}

	if qq == 0 {
		qq = vector.DotUnsafe(yv, yv, dp.vectorLength) + ydf*ydf
	}

	var ppqq TV

	if pp != 0 {
		ppqq = pp * qq
	}

	if ppqq > 0 {
		pq := vector.DotUnsafe(xv, yv, dp.vectorLength) + xdf*ydf
		return 2.0 - 2.0*pq/TV(math.Sqrt(float64(ppqq)))
	}
	return 2.0
}

func (dp *dotProductDistanceImpl[TV]) PreProcess(nodes unsafe.Pointer, nodeCount interfaces.ItemID) {
	// This uses a method from Microsoft Research for transforming inner product spaces to cosine/angular-compatible spaces.
	// (Bachrach et al., 2014, see https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/XboxInnerProduct.pdf)

	// Step one: compute the norm of each vector and store that in its extra dimension (f-1)
	for i := 0; i < int(nodeCount); i++ {
		node := dp.MapNodeToMemory(nodes, interfaces.ItemID(i))
		dn := node.(*DotProductNodeImpl[TV])
		nv := node.GetRawVector()
		d := vector.DotUnsafe(nv, nv, dp.vectorLength)

		var norm TV
		if d >= 0 {
			norm = TV(math.Sqrt(float64(d)))
		}

		dn.dot_factor = norm
		dn.built = false
	}

	// Step two: find the maximum norm
	max_norm := TV(0)

	for i := 0; i < int(nodeCount); i++ {
		node := dp.MapNodeToMemory(nodes, interfaces.ItemID(i))
		df := node.(*DotProductNodeImpl[TV]).dot_factor

		if df > max_norm {
			max_norm = df
		}
	}
	// Step three: set each vector's extra dimension to sqrt(max_norm^2 - norm^2)
	for i := 0; i < int(nodeCount); i++ {
		node := dp.MapNodeToMemory(nodes, interfaces.ItemID(i))
		dn := node.(*DotProductNodeImpl[TV])
		node_norm := dn.dot_factor

		squared_norm_diff := TV(math.Pow(float64(max_norm), 2.0)) - TV(math.Pow(float64(node_norm), 2.0))

		var dot_factor TV
		if squared_norm_diff >= 0 {
			dot_factor = TV(math.Sqrt(float64(squared_norm_diff)))
		}

		dn.SetNorm(max_norm * max_norm)
		dn.dot_factor = dot_factor
	}
}

// InitNode will initialize the node by setting the norm to the value based on the distance type.
func (dp *dotProductDistanceImpl[TV]) InitNode(node interfaces.Node[TV]) {
	dn := node.(*DotProductNodeImpl[TV])
	dn.built = false
	dn.SetNorm(
		vector.DotUnsafe(node.GetRawVector(), node.GetRawVector(), dp.vectorLength) +
			dn.dot_factor*dn.dot_factor,
	)
}

func (dp *dotProductDistanceImpl[TV]) MeanNorm(node interfaces.Node[TV], vectorLength int) TV {
	dn := node.(*DotProductNodeImpl[TV])
	return TV(math.Sqrt(float64(
		vector.DotUnsafe(node.GetRawVector(), node.GetRawVector(), vectorLength) +
			dn.dot_factor*dn.dot_factor,
	)))
}

func (dp *dotProductDistanceImpl[TV]) UpdateMean(
	mean interfaces.Node[TV],
	newNode interfaces.Node[TV],
	norm TV,
	count int,
	vectorLength int,
) {
	mv := mean.GetVector(vectorLength)
	nv := newNode.GetVector(vectorLength)
	c := TV(count)
	next := TV(count + 1)

	for z := 0; z < vectorLength; z++ {
		mv[z] = (mv[z]*c + nv[z]/norm) / next
	}

	md := mean.(*DotProductNodeImpl[TV])
	nd := newNode.(*DotProductNodeImpl[TV])
	md.dot_factor = (md.dot_factor*c + nd.dot_factor/norm) / next
}

func (dp *dotProductDistanceImpl[TV]) PostProcess(nodes unsafe.Pointer, nodeCount interfaces.ItemID) {
	for i := 0; i < int(nodeCount); i++ {
		node := dp.MapNodeToMemory(nodes, interfaces.ItemID(i))
		node.(*DotProductNodeImpl[TV]).built = true
	}
}

func (dp *dotProductDistanceImpl[TV]) Name() string {
	return "dotproduct"
}
