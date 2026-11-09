package engine

import (
	"context"
	"fmt"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/core"
	"github.com/mariotoffia/goannoy/internal/metric"
	"math"
	"os"
)

type Builder struct {
	Config *core.BuildConfig
	layout codec.Layout
	items  []byte
}

func NewBuilder(k codec.Kind, dimension int) (*Builder, error) {
	l, err := codec.New(k, dimension)
	if err != nil {
		return nil, err
	}
	c, err := core.NewBuildConfig(dimension)
	if err != nil {
		return nil, err
	}
	return &Builder{Config: c, layout: l}, nil
}
func (b *Builder) prepare(id goannoy.ItemID) ([]byte, error) {
	if err := b.Config.Err(); err != nil {
		return nil, err
	}
	if id < 0 || id == math.MaxInt32 || int64(id+1)*int64(b.layout.Stride) > int64(math.MaxInt) {
		return nil, goannoy.ErrInvalidID
	}
	size := (int(id) + 1) * b.layout.Stride
	if size > len(b.items) {
		c := b.Config.Resolve().Capacity * b.layout.Stride
		if c < size {
			c = size
		}
		if c > cap(b.items) {
			next := make([]byte, size, max(c, cap(b.items)+cap(b.items)/2))
			copy(next, b.items)
			b.items = next
		} else {
			b.items = b.items[:size]
		}
	}
	node := b.layout.Node(b.items, int(id))
	clear(node)
	codec.PutI32(node, 0, 1)
	return node, nil
}
func (b *Builder) AddFloat(id goannoy.ItemID, v []float32) error {
	if err := b.Config.Err(); err != nil {
		return err
	}
	if err := b.layout.CheckFloats(v); err != nil {
		return err
	}
	n, err := b.prepare(id)
	if err != nil {
		return err
	}
	copy(b.layout.Floats(n, 0), v)
	metric.Init(b.layout, n)
	return nil
}
func (b *Builder) AddWords(id goannoy.ItemID, v []uint64) error {
	if err := b.Config.Err(); err != nil {
		return err
	}
	if err := b.layout.CheckWords(v); err != nil {
		return err
	}
	n, err := b.prepare(id)
	if err != nil {
		return err
	}
	copy(b.layout.WordsAt(n, 0), v)
	return nil
}
func (b *Builder) Build(ctx context.Context) (*Index, error) {
	if path := b.Config.Resolve().Path; path != "" {
		return b.BuildTo(ctx, path)
	}
	data, err := b.build(ctx)
	if err != nil {
		return nil, err
	}
	return FromMapping(&core.Mapping{Data: data}, b.layout)
}
func (b *Builder) BuildTo(ctx context.Context, path string) (*Index, error) {
	if path == "" {
		return nil, goannoy.ErrInvalidOption
	}
	err := core.AtomicWrite(path, func(f *os.File) error {
		if b.Config.Resolve().Path != "" {
			storage := core.NewMutableFile(f)
			_, err := b.buildWithStorage(ctx, storage)
			closeErr := storage.Close()
			if err != nil {
				return err
			}
			return closeErr
		}
		data, err := b.build(ctx)
		if err != nil {
			return err
		}
		_, err = f.Write(data)
		return err
	})
	if err != nil {
		return nil, err
	}
	return Open(path, b.layout.Kind, b.layout.Dimension)
}
func (b *Builder) build(ctx context.Context) ([]byte, error) {
	return b.buildWithStorage(ctx, nil)
}
func (b *Builder) buildWithStorage(ctx context.Context, storage *core.MutableFile) ([]byte, error) {
	if err := b.Config.Err(); err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	l := b.layout
	items := len(b.items) / l.Stride
	if items == 0 {
		return nil, fmt.Errorf("%w: no items", goannoy.ErrInvalidOption)
	}
	var data []byte
	if storage != nil {
		var err error
		data, err = storage.Append(b.items)
		if err != nil {
			return nil, err
		}
	} else {
		data = append([]byte(nil), b.items...)
	}
	if l.Kind == codec.Dot {
		preprocessDot(l, data)
	}
	ids := make([]int32, 0, items)
	for id := 0; id < items; id++ {
		if l.Descendants(data, id) == 1 {
			ids = append(ids, int32(id))
		}
	}
	settings := b.Config.Resolve()
	if storage == nil && settings.Workers > 1 && settings.Trees > 1 {
		data, err := parallelForest(ctx, l, data, ids, settings)
		if err != nil {
			return nil, err
		}
		if l.Kind == codec.Dot {
			for id := 0; id < items; id++ {
				l.Node(data, id)[20] = 1
			}
		}
		return data, nil
	}
	s := treeBuilder{layout: l, data: data, items: items, ctx: ctx, random: core.NewKiss64(settings.Seed), storage: storage}
	roots := make([]int32, 0, settings.Trees)
	for n := 0; n < settings.Trees; n++ {
		root, err := s.tree(ids, true)
		if err != nil {
			return nil, err
		}
		roots = append(roots, root)
	}
	for _, root := range roots {
		copyNode := append([]byte(nil), l.Node(s.data, int(root))...)
		if _, err := s.appendNode(copyNode); err != nil {
			return nil, err
		}
	}
	if l.Kind == codec.Dot {
		for id := 0; id < items; id++ {
			l.Node(s.data, id)[20] = 1
		}
	}
	return s.data, nil
}
func preprocessDot(l codec.Layout, data []byte) {
	count := len(data) / l.Stride
	var maxNorm float32
	for id := 0; id < count; id++ {
		node := l.Node(data, id)
		v := l.Floats(node, 0)
		norm := float32(math.Sqrt(float64(max(metric.Dot(v, v), 0))))
		codec.PutF32(node, 12, norm)
		node[20] = 0
		maxNorm = max(maxNorm, norm)
	}
	for id := 0; id < count; id++ {
		node := l.Node(data, id)
		norm := codec.F32(node, 12)
		mx := maxNorm
		diff := float32(mx*mx) - float32(norm*norm)
		codec.PutF32(node, 12, float32(math.Sqrt(float64(max(diff, 0)))))
		codec.PutF32(node, 16, float32(mx*mx))
	}
}
