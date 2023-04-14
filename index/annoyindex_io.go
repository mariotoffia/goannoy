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

	for i := idx._n_nodes - 1; i >= 0; i-- {

		n := idx.getNode(i)
		k := n.GetNumberOfDescendants()

		if !mset || k == m {
			idx._roots = append(idx._roots, i)
			m = k
			mset = true
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

	if idx.logVerbose {
		fmt.Println("Loaded index to from file", fileName, "with", idx._n_nodes, "nodes")
	}

	idx.batchMaxNNS = -1

	for i := TIX(0); i < idx._n_nodes; i++ {
		nd := idx.getNode(i)

		nDescendants := nd.GetNumberOfDescendants()

		if nDescendants == 1 && i < idx._n_items {
			idx.batchMaxNNS++
		} else if nDescendants <= idx.maxDescendants {
			idx.batchMaxNNS += len(nd.GetChildren())
		}

		if idx.logVerbose {
			fmt.Println(i, ": ", utils.DumpNode(idx.distance, nd))
		}
	}

	if idx.logVerbose {
		fmt.Println("Max NNS:", idx.batchMaxNNS)
	}

	return nil
}
