package engine

import (
	"context"
	"github.com/mariotoffia/goannoy/internal/codec"
	"github.com/mariotoffia/goannoy/internal/core"
	"sync"
)

type treeResult struct {
	nodes []byte
	root  int32
	err   error
}

func parallelForest(ctx context.Context, l codec.Layout, items []byte, ids []int32, settings core.Settings) ([]byte, error) {
	results := make([]treeResult, settings.Trees)
	jobs := make(chan int)
	var wg sync.WaitGroup
	count := len(items) / l.Stride
	for worker := 0; worker < min(settings.Workers, settings.Trees); worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := range jobs {
				// Full slice capacity forces the first append to copy. Item data is read-only.
				s := treeBuilder{layout: l, data: items[:len(items):len(items)], items: count, ctx: ctx, random: core.NewKiss64(settings.Seed + uint64(n))}
				root, err := s.tree(ids, true)
				if err != nil {
					results[n].err = err
					continue
				}
				results[n] = treeResult{nodes: append([]byte(nil), s.data[len(items):]...), root: root}
			}
		}()
	}
	for n := range results {
		jobs <- n
	}
	close(jobs)
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	roots := make([]int32, 0, len(results))
	data := items
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		offset := int32(len(data)/l.Stride - count)
		for id := 0; id < len(r.nodes)/l.Stride; id++ {
			if int(l.Descendants(r.nodes, id)) <= l.LeafCapacity {
				continue
			}
			node := l.Node(r.nodes, id)
			for side, child := range l.Children(r.nodes, id) {
				if int(child) >= count {
					codec.PutI32(node, l.ChildrenOffset+4*side, child+offset)
				}
			}
		}
		data = append(data, r.nodes...)
		roots = append(roots, r.root+offset)
	}
	for _, root := range roots {
		data = append(data, l.Node(data, int(root))...)
	}
	return data, nil
}
