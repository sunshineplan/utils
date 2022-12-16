package loadbalance

import (
	"math/rand"
	"time"
)

var _ LoadBalancer[struct{}] = &random[struct{}]{}

func init() {
	rand.Seed(time.Now().UnixNano())
}

type random[E any] struct {
	items []E
}

func Random[E any](items ...E) (LoadBalancer[E], error) {
	if len(items) == 0 {
		return nil, ErrEmptyLoadBalancer
	}
	return &random[E]{items: items}, nil
}

func RandomWeighted[E any](items ...Weighted[E]) (LoadBalancer[E], error) {
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

func (r *random[E]) Next() E {
	return r.items[rand.Intn(len(r.items))]
}
