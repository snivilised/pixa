package proxy

import (
	"sync"
)

func NewControllerRegistry(shared *SharedControllerInfo) *ControllerRegistry {
	return &ControllerRegistry{
		pool: sync.Pool{
			// see: https://www.sobyte.net/post/2022-03/think-in-sync-pool/
			//
			New: func() interface{} {
				switch shared.Type {
				case ControllerTypeFullEn:
					return &FullController{
						controller: controller{
							shared: shared,
						},
					}

				case ControllerTypeSamplerEn:
					return &SamplerController{
						controller: controller{
							shared: shared,
						},
					}
				}

				panic("undefined controller type")
			},
		},
	}
}

type ControllerRegistry struct {
	pool sync.Pool
}

func (cr *ControllerRegistry) Get() ItemController {
	return cr.pool.Get().(ItemController)
}

func (cr *ControllerRegistry) Put(controller ItemController) {
	controller.Reset()
	cr.pool.Put(controller)
}
