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
	if ctx.Err() != nil {
		return
	}

	var r interface{}
	var err error
	c := make(chan error, 1)
	go func() {
		r, err = ctx.fn(s)
		c <- err
	}()

	select {
	case <-ctx.Done():
		return
	case err := <-c:
		ctx.Lock()
		defer ctx.Unlock()

		if err != nil {
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
