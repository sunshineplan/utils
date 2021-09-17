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

func newFnContext(count int, fn func(interface{}) (interface{}, error)) fnContext {
	ctx, cancel := context.WithCancel(context.Background())
	return fnContext{ctx, cancel, sync.Mutex{}, count, fn}
}

func (ctx *fnContext) run(arg interface{}, rc chan<- interface{}, ec chan<- error) {
	if ctx.Err() != nil {
		return
	}

	var r interface{}
	var err error
	c := make(chan error, 1)
	go func() {
		r, err = ctx.fn(arg)
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

type argContext struct {
	context.Context
	cancel func()

	sync.Mutex

	count int
	arg   interface{}
}

func newArgContext(count int, arg interface{}) argContext {
	ctx, cancel := context.WithCancel(context.Background())
	return argContext{ctx, cancel, sync.Mutex{}, count, arg}
}

func (ctx *argContext) run(fn func(interface{}) (interface{}, error), rc chan<- interface{}, ec chan<- error) {
	if ctx.Err() != nil {
		return
	}

	var r interface{}
	var err error
	c := make(chan error, 1)
	go func() {
		r, err = fn(ctx.arg)
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
