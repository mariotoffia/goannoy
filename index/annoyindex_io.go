package index

import (
	"errors"
	"fmt"
	"os"
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/utils"
)

func (idx *AnnoyIndexImpl[TV]) OnDiskBuild(fileName string) error {
	if idx._n_items > 0 {
		return errors.New("on-disk build must be called before adding items")
	}
	if idx.indexBuilt || idx.indexLoaded {
		return errors.New("on-disk build requires a fresh index")
	}

	alloc, err := newOnDiskBuildAllocator(fileName)
	if err != nil {
		return err
	}

	// Free the old allocator and switch to on-disk.
	if idx.allocator != nil {
		idx.allocator.Free()
	}
	idx.allocator = alloc
	idx.onDiskFile = fileName

	return nil
}

func (idx *AnnoyIndexImpl[TV]) Unbuild() error {
	if idx.indexLoaded {
		return errors.New("can't unbuild a loaded index")
	}
	if !idx.indexBuilt {
		return errors.New("can't unbuild an index that hasn't been built")
	}

	idx._roots = nil
	idx._n_nodes = idx._n_items
	idx.indexBuilt = false
	idx.batchMaxNNS = -1

	return nil
}

func (idx *AnnoyIndexImpl[TV]) Unload() error {
	var err error

	if idx.indexMemory != nil {
		err = idx.indexMemory.Close()
		idx.indexMemory = nil
	}

	idx._nodes = nil
	idx.indexLoaded = false
	idx.indexBuilt = false
	idx._n_items = 0
	idx._n_nodes = 0
	idx._nodes_size = 0
	idx._roots = nil
	idx.batchMaxNNS = -1

	return err
}

// Close implements the io.Closer interface.
func (idx *AnnoyIndexImpl[TV]) Close() error {
	var err error

	if idx.indexMemory != nil {
		err = idx.indexMemory.Close()
		idx.indexMemory = nil
	}

	if idx.allocator != nil {
		idx.allocator.Free()
	}

	idx._nodes = nil
	idx.indexLoaded = false
	idx._n_items = 0
	idx._n_nodes = 0
	idx._nodes_size = 0
	if idx.random != nil {
		idx.random = idx.random.CloneAndReset()
	}
	idx._roots = nil

	return err
}

func (idx *AnnoyIndexImpl[TV]) Prefault() error {
	if p, ok := idx.indexMemory.(interface{ Prefault() error }); ok {
		return p.Prefault()
	}
	return nil
}

func (idx *AnnoyIndexImpl[TV]) Save(fileName string) error {
	if !idx.indexBuilt {
		return fmt.Errorf("can't save an index that hasn't been built")
	}

	// Match upstream Annoy behavior: unlink first so an already-mapped reader
	// keeps the old inode alive instead of observing an in-place truncation.
	_ = os.Remove(fileName)

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	if idx.logVerbose {
		fmt.Println("Saving index to file", fileName, "with", idx._n_nodes, "nodes")
		for i := 0; i < int(idx._n_nodes); i++ {
			n := idx.getNode(interfaces.ItemID(i))
			fmt.Println(i, ": ", utils.DumpNode(idx.distance, n))
		}
	}

	data := unsafe.Slice((*byte)(idx._nodes), int(idx._n_nodes)*idx.nodeSize)

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return idx.Load(fileName)
}

func (idx *AnnoyIndexImpl[TV]) Load(fileName string) error {
	// Close any existing index and free resources.
	idx.Close()

	var err error

	idx.indexMemory, err = idx.indexMemoryAllocator.Open(fileName)
	if err != nil {
		return err
	}

	if idx.indexMemory.Size()%int64(idx.nodeSize) != 0 {
		idx.Close()
		return fmt.Errorf("file size is not a multiple of node size")
	}

	idx._nodes = idx.indexMemory.Ptr()
	idx._roots = nil
	idx._n_nodes = interfaces.ItemID(idx.indexMemory.Size() / int64(idx.nodeSize))

	var (
		mset bool
		m    interfaces.ItemID
	)

	// Scan backwards to discover root nodes. Root copies are appended at the
	// end of the file during Build and share the same n_descendants value
	// (== _n_items).
	for i := int(idx._n_nodes); i > 0; i-- {
		nodeID := interfaces.ItemID(i - 1)
		n := idx.getNode(nodeID)
		k := n.GetNumberOfDescendants()

		if !mset || k == m {
			idx._roots = append(idx._roots, nodeID)
			m = k
			mset = true
		} else {
			break
		}
	}

	idx._n_items = m

	if idx._n_items <= 1 {
		// Single-item indexes are ambiguous because the item leaf and root nodes
		// can both have n_descendants == 1. Keep only non-item nodes.
		filtered := idx._roots[:0]
		for _, r := range idx._roots {
			if r >= idx._n_items {
				filtered = append(filtered, r)
			}
		}
		idx._roots = filtered
	}

	// Upstream Annoy keeps the roots discovered by the backward scan and only
	// removes the duplicated original root that may precede the tail copies.
	if len(idx._roots) > 1 {
		firstChildren := idx.getNode(idx._roots[0]).GetChildren()
		lastChildren := idx.getNode(idx._roots[len(idx._roots)-1]).GetChildren()
		if len(firstChildren) > 0 && len(lastChildren) > 0 &&
			firstChildren[0] == lastChildren[0] {
			idx._roots = idx._roots[:len(idx._roots)-1]
		}
	}

	idx.indexBuilt = true
	idx.indexLoaded = true

	if idx.logVerbose {
		fmt.Println("Loaded index to from file", fileName, "with", idx._n_nodes, "nodes")
		for i := 0; i < int(idx._n_nodes); i++ {
			nd := idx.getNode(interfaces.ItemID(i))
			fmt.Println(i, ": ", utils.DumpNode(idx.distance, nd))
		}
	}

	idx.computeBatchMaxNNS()

	return nil
}
