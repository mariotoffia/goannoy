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
	treesPerThread int, nThreads int,
) {
	if nThreads == -1 {
		nThreads = utils.Max(1, runtime.NumCPU())
	}

	var wg sync.WaitGroup
	wg.Add(nThreads)

	for threadIdx := 0; threadIdx < nThreads; threadIdx++ {
		tpt := treesPerThread
		if treesPerThread != -1 {
			tpt = int(math.Floor(float64(treesPerThread+threadIdx) / float64(nThreads)))
		}

		go func(threadIdx int, treesPerThread int) {

			defer wg.Done()

			builder.ThreadBuild(treesPerThread, threadIdx, p)

		}(threadIdx, tpt)
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
