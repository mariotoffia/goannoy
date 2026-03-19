package index

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/utils"
)

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
