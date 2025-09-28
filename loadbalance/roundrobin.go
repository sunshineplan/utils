package loadbalance

import (
	"sync"

	"github.com/sunshineplan/utils/container"
)

var _ LoadBalancer[any] = &roundrobin[any]{}

type roundrobin[E any] struct {
	sync.RWMutex
	ring *container.Ring[E]
}

func newRoundRobin[E any, Items []E | []Weighted[E]](items Items) *roundrobin[E] {
	if len(items) == 0 {
		panic(ErrEmptyLoadBalancer)
	}
	var ring *container.Ring[E]
	switch items := any(items).(type) {
	case []E:
		ring = container.NewRing[E](len(items))
		for _, i := range items {
			ring.Set(i)
			ring = ring.Next()
		}
	case []Weighted[E]:
		for _, i := range items {
			if i.Weight == 0 {
				continue
			}
			subring := container.NewRing[E](i.Weight)
			for range i.Weight {
				subring.Set(i.Item)
				subring = subring.Next()
			}
			if ring == nil {
				ring = subring
			} else {
				ring = ring.Link(subring).Prev()
			}
		}
		if ring != nil {
			ring = ring.Next()
		}
	}
	return &roundrobin[E]{ring: ring}
}

func RoundRobin[E any](items ...E) LoadBalancer[E] {
	return newRoundRobin[E](items)
}

func WeightedRoundRobin[E any](items ...Weighted[E]) LoadBalancer[E] {
	return newRoundRobin[E](items)
}

func (r *roundrobin[E]) Len() int {
	r.RLock()
	defer r.RUnlock()
	return r.ring.Len()
}

func (r *roundrobin[E]) Next() (next E) {
	r.Lock()
	defer r.Unlock()
	next = r.ring.Value()
	r.ring = r.ring.Next()
	return
}

func (r *roundrobin[E]) Ring() *container.Ring[E] {
	r.RLock()
	defer r.RUnlock()
	return r.ring
}

func (r *roundrobin[E]) Link(s LoadBalancer[E]) LoadBalancer[E] {
	sr := s.Ring()
	r.Lock()
	defer r.Unlock()
	r.ring = r.ring.Prev().Link(sr)
	return r
}

func (r *roundrobin[E]) Unlink(n int) LoadBalancer[E] {
	r.Lock()
	defer r.Unlock()
	r.ring = r.ring.Unlink(n)
	return r
}
