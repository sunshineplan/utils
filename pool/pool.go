package pool

import "sync"

// Pool is a type-safe wrapper around [sync.Pool] that provides
// a generic, concurrency-safe object pool for any type T.
type Pool[T any] struct {
	pool sync.Pool
	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	// It may not be changed concurrently with calls to Get.
	New func() *T
}

// New creates a new Pool that allocates new zero-valued objects
// of type T when none are available.
func New[T any]() *Pool[T] {
	return &Pool[T]{New: func() *T { return new(T) }}
}

// Get selects an arbitrary item from the [Pool], removes it from the
// Pool, and returns it to the caller.
// Get may choose to ignore the pool and treat it as empty.
// Callers should not assume any relation between values passed to [Pool.Put] and
// the values returned by Get.
//
// If Get would otherwise return nil and p.New is non-nil, Get returns
// the result of calling p.New.
func (p *Pool[T]) Get() *T {
	x := p.pool.Get()
	if x == nil {
		if p.New != nil {
			return p.New()
		}
		return nil
	}
	return x.(*T)
}

// Put adds x to the pool.
func (p *Pool[T]) Put(x *T) {
	p.pool.Put(x)
}
