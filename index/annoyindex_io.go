package index

import (
	"fmt"
	"os"
	"unsafe"
)

func (idx *AnnoyIndexImpl[TV, TR]) Save(fileName string) error {
	if !idx.indexBuilt {
		return fmt.Errorf("can't save an index that hasn't been built")
	}

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	data := unsafe.Slice((*byte)(idx._nodes), idx._n_nodes*idx.nodeSize)

	_, err = file.Write(data)

	if err != nil {
		return err
	}

	return idx.Load(fileName)
}

func (idx *AnnoyIndexImpl[TV, TR]) Load(fileName string) error {
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
	idx._n_nodes = int(idx.indexMemory.Size()) / idx.nodeSize

	m := -1

	for i := idx._n_nodes - 1; i >= 0; i++ {

		n := idx.getNode(i)
		k := n.GetNumberOfDescendants()

		if m == -1 || k == m {
			idx._roots = append(idx._roots, i)
			m = k
		} else {
			break
		}
	}

	// hacky fix: since the last root precedes the copy of all roots, delete it
	if len(idx._roots) > 1 {
		fn := idx.getNode(idx._roots[0])
		ln := idx.getNode(idx._roots[len(idx._roots)-1])

		if fn.GetChildren()[0] == ln.GetChildren()[0] {
			idx._roots = idx._roots[:len(idx._roots)-1]
		}
	}

	idx.indexBuilt = true
	idx.indexLoaded = true
	idx._n_items = m

	return nil
}
