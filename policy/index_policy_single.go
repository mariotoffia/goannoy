package policy

import "github.com/mariotoffia/goannoy/interfaces"

type AnnoyIndexSingleThreadedBuildPolicy struct{}

func (p *AnnoyIndexSingleThreadedBuildPolicy) Build(
	builder interfaces.AnnoyIndexBuilder,
	numberOfTrees, nThreads int,
) {
	builder.ThreadBuild(numberOfTrees, 0, p)
}

func (p *AnnoyIndexSingleThreadedBuildPolicy) LockNNodes() {
}

func (p *AnnoyIndexSingleThreadedBuildPolicy) UnlockNNodes() {
}

func (p *AnnoyIndexSingleThreadedBuildPolicy) LockNodes() {
}

func (p *AnnoyIndexSingleThreadedBuildPolicy) UnlockNodes() {
}

func (p *AnnoyIndexSingleThreadedBuildPolicy) LockSharedNodes() {
}

func (p *AnnoyIndexSingleThreadedBuildPolicy) UnlockSharedNodes() {
}

func (p *AnnoyIndexSingleThreadedBuildPolicy) LockRoots() {
}

func (p *AnnoyIndexSingleThreadedBuildPolicy) UnlockRoots() {
}
