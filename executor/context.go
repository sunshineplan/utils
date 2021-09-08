package executor

import (
	"context"
	"sync"
)

type fnContext struct {
	context.Context
	cancel func()

	sync.Mutex

	count int
	fn    func(interface{}) (interface{}, error)
}

func newContext(count int, fn func(interface{}) (interface{}, error)) fnContext {
	ctx, cancel := context.WithCancel(context.Background())
	return fnContext{ctx, cancel, sync.Mutex{}, count, fn}
}

func (ctx *fnContext) run(s interface{}, rc chan<- interface{}, ec chan<- error) {
	c := make(chan error, 1)

	var r interface{}
	var err error
	go func() {
		r, err = ctx.fn(s)
		c <- err
	}()

	select {
	case <-ctx.Done():
		return
	case err := <-c:
		if err != nil {
			ctx.Lock()
			defer ctx.Unlock()

			if ctx.count == 1 {
				rc <- nil
				ec <- err
			}
			ctx.count--
		} else {
			ctx.cancel()

			rc <- r
			ec <- nil
		}
	}
}
