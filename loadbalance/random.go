package loadbalance

import "math/rand/v2"

var _ LoadBalancer[any] = &random[any]{}

type random[E any] struct {
	*roundrobin[E]
}

func Random[E any](items ...E) LoadBalancer[E] {
	return &random[E]{newRoundRobin[E](items)}
}

func WeightedRandom[E any](items ...Weighted[E]) LoadBalancer[E] {
	return &random[E]{newRoundRobin[E](items)}
}

func (r *random[E]) Next() E {
	mu.RLock()
	defer mu.RUnlock()
	r.roundrobin = (*roundrobin[E])(r.Ring().Move(rand.IntN(r.Ring().Len())))
	return **r.Ring().Value()
}
