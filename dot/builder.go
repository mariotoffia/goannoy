package dot

import (
	"context"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/engine"
)

// Builder owns mutable input. Configuration errors are sticky; data errors are not.
type Builder struct{ inner *engine.Builder }

func NewBuilder(dimension int) (*Builder, error) {
	b, err := engine.NewBuilder(codec.Dot, dimension)
	if err != nil {
		return nil, err
	}
	return &Builder{inner: b}, nil
}
func (b *Builder) Trees(n int) *Builder                     { b.inner.Config.Trees(n); return b }
func (b *Builder) Workers(n int) *Builder                   { b.inner.Config.Workers(n); return b }
func (b *Builder) Seed(n uint64) *Builder                   { b.inner.Config.Seed(n); return b }
func (b *Builder) Capacity(n int) *Builder                  { b.inner.Config.Capacity(n); return b }
func (b *Builder) Profile(p goannoy.Profile) *Builder       { b.inner.Config.Profile(p); return b }
func (b *Builder) OnDisk(path string) *Builder              { b.inner.Config.OnDisk(path); return b }
func (b *Builder) Err() error                               { return b.inner.Config.Err() }
func (b *Builder) Add(id goannoy.ItemID, v []float32) error { return b.inner.AddFloat(id, v) }
func (b *Builder) Build(ctx context.Context) (*Index, error) {
	i, err := b.inner.Build(ctx)
	if err != nil {
		return nil, err
	}
	return &Index{inner: i}, nil
}
func (b *Builder) BuildTo(ctx context.Context, path string) (*Index, error) {
	i, err := b.inner.BuildTo(ctx, path)
	if err != nil {
		return nil, err
	}
	return &Index{inner: i}, nil
}
