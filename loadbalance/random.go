package loadbalance

import (
	"math/rand/v2"
	"sync"
)

var _ LoadBalancer[struct{}] = &random[struct{}]{}

type random[E any] struct {
	sync.Mutex
	items []E
	c     chan E
}

func Random[E any](items ...E) (LoadBalancer[E], error) {
	if len(items) == 0 {
		return nil, ErrEmptyLoadBalancer
	}
	return &random[E]{items: items, c: make(chan E, len(items))}, nil
}

func WeightedRandom[E any](items ...Weighted[E]) (LoadBalancer[E], error) {
	var pool []E
	for _, i := range items {
		for n := i.Weight; n > 0; n-- {
			pool = append(pool, i.Item)
		}
	}
	if len(pool) == 0 {
		return nil, ErrEmptyLoadBalancer
	}
	return Random(pool...)
}

func (r *random[E]) load() {
	length := len(r.items)
	var s []int
	for i := 0; i < length; i++ {
		s = append(s, i)
	}
	rand.Shuffle(length, func(i, j int) { s[i], s[j] = s[j], s[i] })
	for _, i := range s {
		r.c <- r.items[i]
	}
}

func (r *random[E]) Next() E {
	r.Lock()
	defer r.Unlock()

	if len(r.c) == 0 {
		r.load()
	}
	return <-r.c
}
