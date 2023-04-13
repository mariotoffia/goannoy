package policy

import (
	"math"
	"runtime"
	"sync"

	"github.com/mariotoffia/goannoy/interfaces"
	"github.com/mariotoffia/goannoy/utils"
)

func Multi() *annoyIndexMultiThreadedBuildPolicy {
	return &annoyIndexMultiThreadedBuildPolicy{}
}

type annoyIndexMultiThreadedBuildPolicy struct {
	nodesMutex  sync.RWMutex
	nNodesMutex sync.Mutex
	rootsMutex  sync.Mutex
}

func (p *annoyIndexMultiThreadedBuildPolicy) Build(
	builder interfaces.AnnoyIndexBuilder,
	numberOfTrees, numberOfWorkers int,
) {
	if numberOfWorkers == -1 {
		numberOfWorkers = int(utils.Max(uint32(1), uint32(runtime.NumCPU())))
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

func (p *annoyIndexMultiThreadedBuildPolicy) LockNNodes() {
	p.nNodesMutex.Lock()
}

func (p *annoyIndexMultiThreadedBuildPolicy) UnlockNNodes() {
	p.nNodesMutex.Unlock()
}

func (p *annoyIndexMultiThreadedBuildPolicy) LockNodes() {
	p.nodesMutex.Lock()
}

func (p *annoyIndexMultiThreadedBuildPolicy) UnlockNodes() {
	p.nodesMutex.Unlock()
}

func (p *annoyIndexMultiThreadedBuildPolicy) LockSharedNodes() {
	p.nodesMutex.RLock()
}

func (p *annoyIndexMultiThreadedBuildPolicy) UnlockSharedNodes() {
	p.nodesMutex.RUnlock()
}

func (p *annoyIndexMultiThreadedBuildPolicy) LockRoots() {
	p.rootsMutex.Lock()
}

func (p *annoyIndexMultiThreadedBuildPolicy) UnlockRoots() {
	p.rootsMutex.Unlock()
}
