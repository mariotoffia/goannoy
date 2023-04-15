package angular

import (
	"testing"

	"github.com/mariotoffia/goannoy/index"
	"github.com/mariotoffia/goannoy/index/memory"
	"github.com/mariotoffia/goannoy/index/policy"
	"github.com/mariotoffia/goannoy/random"
	"github.com/stretchr/testify/assert"
)

func TestGetNnsByVector(t *testing.T) {
	idx := index.NewAnnoyIndexImpl[float32, uint32](
		3,
		random.NewKiss32Random(uint32(0)),
		Distance[float32](uint32(3)),
		policy.SingleWorker(),
		memory.IndexMemoryAllocator(),
		memory.MmapIndexAllocator(),
		false, /*verbose*/
		0,
	)

	defer idx.Close()

	idx.AddItem(0, []float32{0, 0, 1})
	idx.AddItem(1, []float32{0, 1, 0})
	idx.AddItem(2, []float32{1, 0, 0})
	idx.Build(10, -1)

	ctx := idx.CreateContext()

	result, _ := idx.GetNnsByVector([]float32{3, 2, 1}, 3, -1, ctx)
	assert.Equal(t, []int{2, 1, 0}, result)

	result, _ = idx.GetNnsByVector([]float32{1, 2, 3}, 3, -1, ctx)
	assert.Equal(t, []int{0, 1, 2}, result)

	result, _ = idx.GetNnsByVector([]float32{2, 0, 1}, 3, -1, ctx)
	assert.Equal(t, []int{2, 0, 1}, result)
}
