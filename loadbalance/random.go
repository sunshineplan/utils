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
	r.RLock()
	defer r.RUnlock()
	r.roundrobin.ring = r.ring.Move(rand.IntN(r.Len()))
	return r.Ring().Value()
}
