package loadbalance

import "sync/atomic"

var _ LoadBalancer[struct{}] = &roundrobin[struct{}]{}

type roundrobin[E any] struct {
	items []E
	next  uint32
}

func RoundRobin[E any](items ...E) (LoadBalancer[E], error) {
	if len(items) == 0 {
		return nil, ErrEmptyLoadBalancer
	}
	return &roundrobin[E]{items: items}, nil
}

func WeightedRoundRobin[E any](items ...Weighted[E]) (LoadBalancer[E], error) {
	var pool []E
	for _, i := range items {
		for n := i.Weight; n > 0; n-- {
			pool = append(pool, i.Item)
		}
	}
	if len(pool) == 0 {
		return nil, ErrEmptyLoadBalancer
	}
	return &roundrobin[E]{items: pool}, nil
}

func (r *roundrobin[E]) Next() E {
	n := atomic.AddUint32(&r.next, 1)
	return r.items[(int(n)-1)%len(r.items)]
}
