package pool

import "sync"

type Pool[T any] struct {
	p *sync.Pool
}

func New[T any]() *Pool[T] {
	return &Pool[T]{&sync.Pool{New: func() any { return new(T) }}}
}

func (p *Pool[T]) Get() *T {
	return p.p.Get().(*T)
}

func (p *Pool[T]) Put(x *T) {
	p.p.Put(x)
}
