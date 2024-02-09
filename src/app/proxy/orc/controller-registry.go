package orc

import (
	"sync"

	"github.com/snivilised/pixa/src/app/proxy/common"
)

func NewRegistry(session *common.SessionControllerInfo, configs *common.Configs) *ControllerRegistry {
	return &ControllerRegistry{
		pool: sync.Pool{
			// see: https://www.sobyte.net/post/2022-03/think-in-sync-pool/
			//
			New: func() interface{} {
				return New(session, configs)
			},
		},
	}
}

type ControllerRegistry struct {
	pool sync.Pool
}

func (r *ControllerRegistry) Get() common.ItemController {
	return r.pool.Get().(common.ItemController)
}

func (r *ControllerRegistry) Put(controller common.ItemController) {
	controller.Reset()
	r.pool.Put(controller)
}
