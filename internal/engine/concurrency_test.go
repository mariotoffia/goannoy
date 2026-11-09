package engine

import (
	"bytes"
	"context"
	"errors"
	goannoy "github.com/mariotoffia/goannoy"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/core"
	"sync"
	"testing"
)

func TestParallelBuildUsesIndependentSeeds(t *testing.T) {
	b, _ := NewBuilder(codec.Angular, 3)
	b.Config.Trees(4).Seed(42).Workers(1)
	for n := 0; n < 64; n++ {
		if e := b.AddFloat(goannoy.ItemID(n), fixtureFloat(n)); e != nil {
			t.Fatal(e)
		}
	}
	serial, e := b.build(nil)
	if e != nil {
		t.Fatal(e)
	}
	b.Config.Workers(2)
	parallel, e := b.build(nil)
	if e != nil {
		t.Fatal(e)
	}
	if bytes.Equal(serial, parallel) {
		t.Fatal("worker setting ignored; independent tree seeds not used")
	}
	again, e := b.build(nil)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(parallel, again) {
		t.Fatal("parallel serialization is nondeterministic")
	}
	i, e := FromMapping(&core.Mapping{Data: parallel}, b.layout)
	if e != nil {
		t.Fatal(e)
	}
	defer i.Close()
}

func TestBuildCancellation(t *testing.T) {
	b, _ := NewBuilder(codec.Angular, 3)
	b.Config.Workers(2)
	if e := b.AddFloat(0, []float32{1, 0, 0}); e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := b.Build(ctx); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	i, e := b.Build(nil)
	if e != nil {
		t.Fatal(e)
	}
	defer i.Close()
}

func TestIndexLifecycleStateMachine(t *testing.T) {
	i, e := Open("../../compat/fixtures/angular.ann", codec.Angular, 3)
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	for n := 0; n < 16; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := i.NewSearcher()
			dst := make([]goannoy.Result, 10)
			for j := 0; j < 100; j++ {
				_, e := s.SearchFloat([]float32{1, 2, 3}, dst, -1)
				if e != nil && !errors.Is(e, goannoy.ErrClosed) {
					t.Error(e)
				}
			}
		}()
	}
	if e := i.Close(); e != nil {
		t.Fatal(e)
	}
	wg.Wait()
	if e := i.Close(); e != nil {
		t.Fatal(e)
	}
	if _, e := i.NewSearcher().SearchFloat([]float32{1, 2, 3}, nil, -1); !errors.Is(e, goannoy.ErrClosed) {
		t.Fatal(e)
	}
}
