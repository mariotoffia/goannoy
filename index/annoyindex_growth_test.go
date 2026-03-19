package index

import (
	"testing"

	"github.com/mariotoffia/goannoy/distance/angular"
	"github.com/mariotoffia/goannoy/index/memory"
	"github.com/mariotoffia/goannoy/index/policy"
	gorandom "github.com/mariotoffia/goannoy/random"
	"github.com/stretchr/testify/require"
)

func TestAnnoyReallocationUsesUpstreamFactor(t *testing.T) {
	idx := New[float32](
		gorandom.NewKiss32Random(1),
		angular.Distance[float32](3),
		policy.SingleWorker(),
		memory.GoGCIndexAllocator(),
		memory.MmapIndexAllocator(),
		nil,
		false,
		0,
	).(*AnnoyIndexImpl[float32])
	defer idx.Close()

	require.Equal(t, 0, idx._nodes_size)

	idx.allocateSize(1, nil)
	require.Equal(t, 1, idx._nodes_size)

	idx.allocateSize(2, nil)
	require.Equal(t, 2, idx._nodes_size)

	idx.allocateSize(3, nil)
	require.Equal(t, 3, idx._nodes_size)

	idx.allocateSize(4, nil)
	require.Equal(t, 5, idx._nodes_size)
}
