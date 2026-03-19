package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddItemNegativeIDReturnsError(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	err := idx.AddItem(-1, []float32{1, 0, 0})
	require.EqualError(t, err, "negative item id: -1")
}

func TestGetItemNegativeIDPanics(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{1, 0, 0}))
	require.NoError(t, idx.Build(2, 1))

	assert.PanicsWithValue(t, "GetItem: negative item id -1", func() {
		idx.GetItem(-1)
	})
}

func TestGetDistanceNegativeIDPanics(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{1, 0, 0}))
	require.NoError(t, idx.AddItem(1, []float32{0, 1, 0}))
	require.NoError(t, idx.Build(2, 1))

	assert.PanicsWithValue(t, "GetDistance: negative item id -1", func() {
		idx.GetDistance(-1, 0)
	})
}

func TestGetNnsByItemNegativeIDPanics(t *testing.T) {
	idx := angularIdx(3)
	defer idx.Close()

	require.NoError(t, idx.AddItem(0, []float32{1, 0, 0}))
	require.NoError(t, idx.Build(2, 1))

	ctx := idx.CreateContext()
	assert.PanicsWithValue(t, "GetNnsByItem: negative item id -1", func() {
		idx.GetNnsByItem(-1, 1, -1, ctx)
	})
}
