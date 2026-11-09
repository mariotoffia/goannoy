package engine

import (
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/core"
	"github.com/mariotoffia/goannoy/internal/metric"
)

// Spec selects an engine at construction. Static specs let the linker remove
// unused metric entry points. Search dispatch occurs once per operation.
type Spec struct {
	kind   codec.Kind
	raw    metric.RawFunc
	split  func(codec.Layout, []byte, []int32, *core.Kiss64, []byte)
	search func(*Searcher, []goannoy.Result, int) (int, error)
}

var AngularSpec = Spec{codec.Angular, metric.RawAngular, metric.SplitAngular, (*Searcher).searchAngular}
var DotSpec = Spec{codec.Dot, metric.RawDot, metric.SplitDot, (*Searcher).searchDot}
var EuclideanSpec = Spec{codec.Euclidean, metric.RawEuclidean, metric.SplitEuclidean, (*Searcher).searchEuclidean}
var ManhattanSpec = Spec{codec.Manhattan, metric.RawManhattan, metric.SplitManhattan, (*Searcher).searchManhattan}
var HammingSpec = Spec{codec.Hamming, metric.RawHamming, metric.SplitHamming, (*Searcher).searchHamming}

func specForKind(k codec.Kind) Spec {
	switch k {
	case codec.Angular:
		return AngularSpec
	case codec.Dot:
		return DotSpec
	case codec.Euclidean:
		return EuclideanSpec
	case codec.Manhattan:
		return ManhattanSpec
	default:
		return HammingSpec
	}
}
