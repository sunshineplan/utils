package pool

import "sync"

type Pool[T any] struct {
	p *sync.Pool
}

func New[T any]() *Pool[T] {
	return NewFunc(func() *T { return new(T) })
}

func NewFunc[T any](fn func() *T) *Pool[T] {
	return &Pool[T]{&sync.Pool{New: func() any { return fn() }}}
}

func (p *Pool[T]) Get() *T {
	return p.p.Get().(*T)
}

func (p *Pool[T]) Put(x *T) {
	p.p.Put(x)
}
