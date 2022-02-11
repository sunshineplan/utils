package executor

import (
	"context"
	"sync"
)

type key int

const (
	fnKey key = iota + 1
	argKey
)

type Context struct {
	context.Context
	cancel context.CancelFunc

	mu sync.Mutex

	count int
}

func newContext(count int, key key, value interface{}) *Context {
	ctx, cancel := context.WithCancel(context.Background())
	if value != nil {
		ctx = context.WithValue(ctx, key, value)
	}
	return &Context{ctx, cancel, sync.Mutex{}, count}
}

func (ctx *Context) run(executor func(chan<- interface{}, chan<- error), rc chan<- interface{}, ec chan<- error) {
	if ctx.Err() != nil {
		return
	}

	r := make(chan interface{}, 1)
	c := make(chan error, 1)
	go executor(r, c)

	select {
	case <-ctx.Done():
		return
	case err := <-c:
		ctx.mu.Lock()
		defer ctx.mu.Unlock()

		if err != nil {
			if ctx.count == 1 {
				rc <- nil
				ec <- err
			}
			ctx.count--
		} else {
			ctx.cancel()

			rc <- <-r
			ec <- nil
		}
	}
}

func (ctx *Context) runArg(arg interface{}, rc chan<- interface{}, ec chan<- error) {
	ctx.run(func(c1 chan<- interface{}, c2 chan<- error) {
		r, err := (ctx.Value(fnKey).(func(interface{}) (interface{}, error)))(arg)
		c1 <- r
		c2 <- err
	}, rc, ec)
}

func (ctx *Context) runFn(fn func(interface{}) (interface{}, error), rc chan<- interface{}, ec chan<- error) {
	ctx.run(func(c1 chan<- interface{}, c2 chan<- error) {
		r, err := fn(ctx.Value(argKey))
		c1 <- r
		c2 <- err
	}, rc, ec)
}
