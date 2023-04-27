package dotproduct

import (
	"math"
	"unsafe"

	"github.com/mariotoffia/goannoy/distance"
	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/vector"
)

type dotProductDistanceImpl[TV interfaces.VectorType, TIX interfaces.IndexTypes] struct {
	nodeSize       TIX
	maxNumChildren TIX
	vectorLength   TIX
}

// Distance creates a new angular distance implementation.
func Distance[TV interfaces.VectorType, TIX interfaces.IndexTypes](
	vectorLength TIX,
) *dotProductDistanceImpl[TV, TIX] {

	n := DotProductNodeImpl[TV, TIX]{}

	ad := &dotProductDistanceImpl[TV, TIX]{
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

func (dp *dotProductDistanceImpl[TV, TIX]) VectorLength() TIX {
	return dp.vectorLength
}

func (dp *dotProductDistanceImpl[TV, TIX]) MaxNumChildren() TIX {
	return dp.maxNumChildren
}

func (dp *dotProductDistanceImpl[TV, TIX]) NodeSize() TIX {
	return dp.nodeSize
}

func (dp *dotProductDistanceImpl[TV, TIX]) MapNodeToMemory(
	mem unsafe.Pointer,
	itemIndex TIX,
) interfaces.Node[TV, TIX] {
	pos := unsafe.Add(mem, itemIndex*dp.nodeSize)

	return (*DotProductNodeImpl[TV, TIX])(pos)
}

func (dp *dotProductDistanceImpl[TV, TIX]) CreateSplit(
	nodes []interfaces.Node[TV, TIX],
	nodeSize TIX,
	random interfaces.Random[TIX],
	n interfaces.Node[TV, TIX],
) {
	// Allocate memory for two nodes, and use them as temporary nodes
	p_mem := make([]byte, nodeSize)
	q_mem := make([]byte, nodeSize)

	p := (*DotProductNodeImpl[TV, TIX])(unsafe.Pointer(unsafe.SliceData(p_mem)))
	q := (*DotProductNodeImpl[TV, TIX])(unsafe.Pointer(unsafe.SliceData(q_mem)))

	distance.TwoMeans[TV, TIX](nodes, dp.vectorLength, random, true, p, q, dp)

	nv := n.GetVector(dp.vectorLength)
	qv := q.GetVector(dp.vectorLength)
	pv := p.GetVector(dp.vectorLength)

	for z := TIX(0); z < dp.vectorLength; z++ {
		nv[z] = pv[z] - qv[z]
	}

	dp.Normalize(n)
}

func (dp *dotProductDistanceImpl[TV, TIX]) Normalize(node interfaces.Node[TV, TIX]) {
	raw := node.GetRawVector()
	norm := TV(vector.GetNormUnsafe(raw, dp.vectorLength))

	if norm > 0 {
		ptr := unsafe.Pointer(raw)
		size := TIX(unsafe.Sizeof(TV(0)))

		for i := TIX(0); i < dp.vectorLength; i++ {
			f := (*TV)(unsafe.Pointer(unsafe.Add(ptr, i*size)))
			*f /= norm
		}

		node.(*DotProductNodeImpl[TV, TIX]).dot_factor /= norm
	}
}

func (dp *dotProductDistanceImpl[TV, TIX]) Margin(n interfaces.Node[TV, TIX], y []TV) TV {
	if len(y) == 0 {
		panic("y is empty")
	}

	df := n.(*DotProductNodeImpl[TV, TIX]).dot_factor

	return vector.DotUnsafe(
		n.GetRawVector(),
		(*TV)(unsafe.Pointer(unsafe.SliceData(y))),
		dp.vectorLength,
	) + (df * df)
}

func (dp *dotProductDistanceImpl[TV, TIX]) Side(
	n interfaces.Node[TV, TIX],
	y []TV,
	random interfaces.Random[TIX],
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

func (dp *dotProductDistanceImpl[TV, _]) NormalizedDistance(distance TV) TV {
	return -distance
}

func (dp *dotProductDistanceImpl[TV, TIX]) PQDistance(distance, margin TV, side interfaces.Side) TV {
	if side == interfaces.SideLeft {
		margin = -margin
	}
	return TV(math.Min(float64(distance), float64(margin)))
}

func (dp *dotProductDistanceImpl[TV, TIX]) PQInitialValue() TV {
	return TV(math.Inf(1))
}

func (dp *dotProductDistanceImpl[TV, TIX]) Distance(x interfaces.Node[TV, TIX], y interfaces.Node[TV, TIX]) TV {
	pp := x.GetNorm()
	qq := y.GetNorm()
	xv := x.GetRawVector()
	yv := y.GetRawVector()

	if pp == 0 {
		pp = vector.DotUnsafe(xv, xv, dp.vectorLength)
	}

	if qq == 0 {
		qq = vector.DotUnsafe(yv, yv, dp.vectorLength)
	}

	var ppqq TV

	if pp != 0 {
		ppqq = pp * qq
	}

	if ppqq > 0 {
		pq := vector.DotUnsafe(xv, yv, dp.vectorLength)
		return 2.0 - 2.0*pq/TV(math.Sqrt(float64(ppqq)))
	}
	return 2.0
}

func (dp *dotProductDistanceImpl[TV, TIX]) PreProcess(nodes unsafe.Pointer, node_count TIX) {
	// This uses a method from Microsoft Research for transforming inner product spaces to cosine/angular-compatible spaces.
	// (Bachrach et al., 2014, see https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/XboxInnerProduct.pdf)

	// Step one: compute the norm of each vector and store that in its extra dimension (f-1)
	for i := TIX(0); i < node_count; i++ {
		node := dp.MapNodeToMemory(nodes, i)
		nv := node.GetRawVector()
		d := vector.DotUnsafe(nv, nv, dp.vectorLength)

		var norm TV
		if d >= 0 {
			norm = TV(math.Sqrt(float64(d)))
		}

		node.(*DotProductNodeImpl[TV, TIX]).dot_factor = norm
	}

	// Step two: find the maximum norm
	max_norm := TV(0)

	for i := TIX(0); i < node_count; i++ {
		node := dp.MapNodeToMemory(nodes, i)
		df := node.(*DotProductNodeImpl[TV, TIX]).dot_factor

		if df > max_norm {
			max_norm = df
		}
	}
	// Step three: set each vector's extra dimension to sqrt(max_norm^2 - norm^2)
	for i := TIX(0); i < node_count; i++ {
		node := dp.MapNodeToMemory(nodes, i)
		node_norm := node.(*DotProductNodeImpl[TV, TIX]).dot_factor

		squared_norm_diff := TV(math.Pow(float64(max_norm), 2.0)) - TV(math.Pow(float64(node_norm), 2.0))

		var dot_factor TV
		if squared_norm_diff >= 0 {
			dot_factor = TV(math.Sqrt(float64(squared_norm_diff)))
		}

		node.(*DotProductNodeImpl[TV, TIX]).dot_factor = dot_factor
	}
}

// InitNode will initialize the node by setting the norm to the value based on the distance type.
func (dp *dotProductDistanceImpl[TV, TIX]) InitNode(node interfaces.Node[TV, TIX]) {
	// DO NOTHING
}

func (dp *dotProductDistanceImpl[TV, TIX]) Name() string {
	return "angular"
}
