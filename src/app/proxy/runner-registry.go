package proxy

import (
	"sync"
)

func NewRunnerRegistry(shared *SharedRunnerInfo) *RunnerRegistry {
	return &RunnerRegistry{
		pool: sync.Pool{
			// see: https://www.sobyte.net/post/2022-03/think-in-sync-pool/
			//
			New: func() interface{} {
				switch shared.Type {
				case RunnerTypeFullEn:
					return &FullRunner{
						baseRunner: baseRunner{
							shared: shared,
						},
					}

				case RunnerTypeSamplerEn:
					return &SamplerRunner{
						baseRunner: baseRunner{
							shared: shared,
						},
					}
				}

				panic("undefined runner type")
			},
		},
	}
}

type RunnerRegistry struct {
	pool sync.Pool
}

func (rr *RunnerRegistry) Get() ItemRunner {
	return rr.pool.Get().(ItemRunner)
}

func (rr *RunnerRegistry) Put(runner ItemRunner) {
	runner.Reset()
	rr.pool.Put(runner)
}
