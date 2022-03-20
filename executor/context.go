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

type fn[T any] interface {
	func(T) (interface{}, error)
}

type Context[T any] struct {
	context.Context
	cancel context.CancelFunc

	mu sync.Mutex

	count int
}

func newContext[T any](ctx context.Context, count int) *Context[T] {
	ctx, cancel := context.WithCancel(ctx)
	return &Context[T]{ctx, cancel, sync.Mutex{}, count}
}

func fnContext[T any, Fn fn[T]](count int, fn Fn) *Context[T] {
	return newContext[T](context.WithValue(context.Background(), fnKey, fn), count)
}

func argContext[T any](count int, arg T) *Context[T] {
	return newContext[T](context.WithValue(context.Background(), argKey, arg), count)
}

func (ctx *Context[T]) run(executor func(chan<- interface{}, chan<- error), rc chan<- interface{}, ec chan<- error) {
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
			if ctx.count <= 1 {
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

func (ctx *Context[T]) runArg(arg T, rc chan<- interface{}, ec chan<- error) {
	ctx.run(func(c1 chan<- interface{}, c2 chan<- error) {
		r, err := (ctx.Value(fnKey).(func(T) (interface{}, error)))(arg)
		c1 <- r
		c2 <- err
	}, rc, ec)
}

func (ctx *Context[T]) runFn(fn func(T) (interface{}, error), rc chan<- interface{}, ec chan<- error) {
	ctx.run(func(c1 chan<- interface{}, c2 chan<- error) {
		r, err := fn(ctx.Value(argKey).(T))
		c1 <- r
		c2 <- err
	}, rc, ec)
}
