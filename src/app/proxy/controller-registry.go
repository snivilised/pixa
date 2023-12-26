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
				return &controller{
					shared:  shared,
					private: &privateControllerInfo{},
				}
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
