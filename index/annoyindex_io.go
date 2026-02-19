package index

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/mariotoffia/goannoy/utils"
)

func (idx *AnnoyIndexImpl[TV, TIX]) Save(fileName string) error {
	if !idx.indexBuilt {
		return fmt.Errorf("can't save an index that hasn't been built")
	}

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	if idx.logVerbose {
		fmt.Println("Saving index to file", fileName, "with", idx._n_nodes, "nodes")
		for i := TIX(0); i < idx._n_nodes; i++ {
			n := idx.getNode(i)
			fmt.Println(i, ": ", utils.DumpNode(idx.distance, n))
		}
	}

	data := unsafe.Slice((*byte)(idx._nodes), idx._n_nodes*idx.nodeSize)

	_, err = file.Write(data)

	if err != nil {
		return err
	}

	return idx.Load(fileName)
}

func (idx *AnnoyIndexImpl[TV, TIX]) Load(fileName string) error {
	// Close any existing index and free resources
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
	idx._n_nodes = TIX(idx.indexMemory.Size()) / idx.nodeSize

	var (
		mset bool
		m    TIX
	)

	// Scan backwards to discover root nodes. Root copies are appended at the
	// end of the file during Build and share the same n_descendants value
	// (== _n_items). Use i > 0 to avoid unsigned underflow with TIX.
	for i := idx._n_nodes; i > 0; i-- {
		n := idx.getNode(i - 1)
		k := n.GetNumberOfDescendants()

		if !mset || k == m {
			idx._roots = append(idx._roots, i-1)
			m = k
			mset = true
		} else {
			break
		}
	}

	idx._n_items = m

	// Filter out item nodes that were incorrectly included as roots.
	// This happens when _n_items == 1 because item leaf nodes and root
	// nodes both have n_descendants == 1.
	var filtered []TIX
	for _, r := range idx._roots {
		if r >= idx._n_items {
			filtered = append(filtered, r)
		}
	}

	// Deduplicate: Build appends copies of the original root nodes at the
	// end of the file. The backward scan picks up both copies (higher indices)
	// and originals (lower indices). We must keep originals over copies because
	// root copies may have incomplete data (CopyNode uses vectorLength, not
	// nodeSize). Since filtered is ordered highest-index-first, we iterate
	// and let later entries (originals) overwrite earlier ones (copies).
	type childKey struct{ c0, c1 TIX }
	rootByKey := make(map[childKey]TIX)
	var keyOrder []childKey

	for _, r := range filtered {
		nd := idx.getNode(r)
		children := nd.GetChildren()

		var key childKey
		if len(children) >= 2 {
			key = childKey{children[0], children[1]}
		} else if len(children) == 1 {
			key = childKey{children[0], 0}
		}

		if _, exists := rootByKey[key]; !exists {
			keyOrder = append(keyOrder, key)
		}
		rootByKey[key] = r // last write wins: originals overwrite copies
	}

	idx._roots = idx._roots[:0]
	for _, key := range keyOrder {
		idx._roots = append(idx._roots, rootByKey[key])
	}

	idx.indexBuilt = true
	idx.indexLoaded = true

	if idx.logVerbose {
		fmt.Println("Loaded index to from file", fileName, "with", idx._n_nodes, "nodes")
	}

	if idx.logVerbose {
		for i := TIX(0); i < idx._n_nodes; i++ {
			nd := idx.getNode(i)
			fmt.Println(i, ": ", utils.DumpNode(idx.distance, nd))
		}
	}

	idx.computeBatchMaxNNS()

	return nil
}
