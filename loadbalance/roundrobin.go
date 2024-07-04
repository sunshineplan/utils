package loadbalance

import "github.com/sunshineplan/utils/cache"

var _ LoadBalancer[any] = &roundrobin[any]{}

type roundrobin[E any] cache.Ring[*E]

func newRoundRobin[E any, Items []E | []Weighted[E]](items Items) *roundrobin[E] {
	if len(items) == 0 {
		panic(ErrEmptyLoadBalancer)
	}
	var root *roundrobin[E]
	switch items := any(items).(type) {
	case []E:
		for _, i := range items {
			r := cache.NewRing[*E](1)
			r.Set(&i)
			if root == nil {
				root = (*roundrobin[E])(r)
			} else {
				root.Ring().Link(r)
				root = (*roundrobin[E])(root.Ring().Next())
			}
		}
		if root != nil {
			root = (*roundrobin[E])(root.Ring().Next())
		}
	case []Weighted[E]:
		for _, i := range items {
			if i.Weight == 0 {
				continue
			}
			item := &i.Item
			r := cache.NewRing[*E](i.Weight)
			for range r.Len() {
				r.Set(item)
				r = r.Next()
			}
			if root == nil {
				root = (*roundrobin[E])(r)
			} else {
				root.Ring().Link(r)
				root = (*roundrobin[E])(root.Ring().Next())
			}
		}
		if root != nil {
			root = (*roundrobin[E])(root.Ring().Next())
		}
	}
	return root
}

func RoundRobin[E any](items ...E) LoadBalancer[E] {
	return newRoundRobin[E](items)
}

func WeightedRoundRobin[E any](items ...Weighted[E]) LoadBalancer[E] {
	return newRoundRobin[E](items)
}

func (r *roundrobin[E]) Len() int {
	mu.RLock()
	defer mu.RUnlock()
	return r.Ring().Len()
}

func (r *roundrobin[E]) Next() (next E) {
	mu.Lock()
	defer mu.Unlock()
	v := **r.Ring().Value()
	*r = *(*roundrobin[E])(r.Ring().Next())
	return v
}

func (r *roundrobin[E]) Ring() *cache.Ring[*E] {
	return (*cache.Ring[*E])(r)
}

func (r *roundrobin[E]) Link(s LoadBalancer[E]) LoadBalancer[E] {
	mu.Lock()
	defer mu.Unlock()
	return (*roundrobin[E])(r.Ring().Link(s.Ring()))
}

func (r *roundrobin[E]) Unlink(n int) LoadBalancer[E] {
	mu.Lock()
	defer mu.Unlock()
	return (*roundrobin[E])(r.Ring().Unlink(n))
}
