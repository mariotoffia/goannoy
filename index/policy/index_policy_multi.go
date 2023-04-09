package policy

import (
	"math"
	"runtime"
	"sync"

	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/utils"
)

type AnnoyIndexMultiThreadedBuildPolicy struct {
	nodesMutex  sync.RWMutex
	nNodesMutex sync.Mutex
	rootsMutex  sync.Mutex
}

func (p *AnnoyIndexMultiThreadedBuildPolicy) Build(
	builder interfaces.AnnoyIndexBuilder,
	numberOfTrees, numberOfWorkers int,
) {
	if numberOfWorkers == -1 {
		numberOfWorkers = utils.Max(1, runtime.NumCPU())
	}

	var wg sync.WaitGroup
	wg.Add(numberOfWorkers)

	for workerIdx := 0; workerIdx < numberOfWorkers; workerIdx++ {
		numTreesPerWorker := numberOfTrees
		if numberOfTrees != -1 {
			numTreesPerWorker = int(math.Floor(float64(numberOfTrees+workerIdx) / float64(numberOfWorkers)))
		}

		go func(workerIdx int, treesPerWorker int) {

			defer wg.Done()

			builder.ThreadBuild(treesPerWorker, workerIdx, p)

		}(workerIdx, numTreesPerWorker)
	}

	wg.Wait()
}

func (p *AnnoyIndexMultiThreadedBuildPolicy) LockNNodes() {
	p.nNodesMutex.Lock()
}

func (p *AnnoyIndexMultiThreadedBuildPolicy) UnlockNNodes() {
	p.nNodesMutex.Unlock()
}

func (p *AnnoyIndexMultiThreadedBuildPolicy) LockNodes() {
	p.nodesMutex.Lock()
}

func (p *AnnoyIndexMultiThreadedBuildPolicy) UnlockNodes() {
	p.nodesMutex.Unlock()
}

func (p *AnnoyIndexMultiThreadedBuildPolicy) LockSharedNodes() {
	p.nodesMutex.RLock()
}

func (p *AnnoyIndexMultiThreadedBuildPolicy) UnlockSharedNodes() {
	p.nodesMutex.RUnlock()
}

func (p *AnnoyIndexMultiThreadedBuildPolicy) LockRoots() {
	p.rootsMutex.Lock()
}

func (p *AnnoyIndexMultiThreadedBuildPolicy) UnlockRoots() {
	p.rootsMutex.Unlock()
}
