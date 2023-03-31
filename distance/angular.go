package distance

import (
	"math"
	"math/rand"

	"github.com/mariotoffia/goannoy/amath"
	"github.com/mariotoffia/goannoy/random"
)

// AngularNodeImpl is a `NodeImpl` to be used with `AngularDistance`
type AngularNodeImpl[TV VectorType] struct {
	NodeImpl[TV]
	children []int32
}

type AngularDistance[TV VectorType, TR random.RandomTypes] struct{}

func (a *AngularDistance[TV, TR]) NewNode(vectorLength int) *AngularNodeImpl[TV] {
	return &AngularNodeImpl[TV]{
		NodeImpl: NodeImpl[TV]{
			v: make([]TV, vectorLength),
		},
		children: make([]int32, 2),
	}

}

// Distance implements `Distance.Distance`
func (a *AngularDistance[TV, TR]) Distance(x, y Node[TV]) TV {
	pp := x.GetNorm()

	xv := x.GetVector()
	yv := y.GetVector()

	if pp == 0 {
		pp = amath.Dot(xv, xv)
	}

	qq := y.GetNorm()
	if qq == 0 {
		qq = amath.Dot(yv, yv)
	}

	pq := amath.Dot(xv, yv)
	ppqq := pp * qq

	if ppqq > 0 {
		return 2.0 - 2.0*pq/TV(math.Sqrt(float64(ppqq)))
	}
	return 2.0
}

// CopyNode implements `Distance.CopyNode`
func (a *AngularDistance[TV, TR]) CopyNode(dst, src Node[TV]) {
	srcVector := src.GetVector()
	destVector := make([]TV, len(srcVector))

	copy(destVector, srcVector)

	dst.SetVector(destVector)
	dst.SetNorm(src.GetNorm())
}

// Normalize implements `Distance.Normalize`
func (a *AngularDistance[TV, TR]) Normalize(node Node[TV]) {
	v := node.GetVector()

	norm := amath.GetNorm(v)

	if norm > 0 {
		l := len(v)
		for i := 0; i < l; i++ {
			v[i] /= norm
		}
	}
}

// InitNode implements `Distance.InitNode`
func (a *AngularDistance[TV, TR]) InitNode(node Node[TV]) {
	v := node.GetVector()

	node.SetNorm(amath.Dot(v, v))
}

func (a *AngularDistance[TV, TR]) Margin(n *AngularNodeImpl[TV], y []TV, f int) TV {
	return amath.Dot(n.v, y)
}

func (a *AngularDistance[TV, TR]) Side(n *AngularNodeImpl[TV], y []TV, f int, random *rand.Rand) bool {
	dotProduct := a.Margin(n, y, f)
	if dotProduct != 0 {
		return dotProduct > 0
	}
	return random.Intn(2) == 1
}

func (a *AngularDistance[TV, TR]) CreateSplit(
	nodes []Node[TV], vectorLength,
	s int,
	random random.Random[TR],
	n *AngularNodeImpl[TV]) {
	//
	p := a.NewNode(vectorLength)
	q := a.NewNode(vectorLength)

	twoMeans[TV](nodes, vectorLength, random, true, p, q, a)

	for z := 0; z < vectorLength; z++ {
		n.v[z] = p.v[z] - q.v[z]
	}

	a.Normalize(n)
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
