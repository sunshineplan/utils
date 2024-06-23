package loadbalance

import (
	"sync/atomic"

	"github.com/sunshineplan/utils/cache"
)

var _ LoadBalancer[any] = &roundrobin[any]{}

type roundrobin[E any] struct {
	m    *cache.Map[[2]int64, *Weighted[E]]
	n    int64
	next atomic.Int64
}

func newRoundRobin[E any, Items []E | []*Weighted[E]](items Items) (*roundrobin[E], error) {
	if len(items) == 0 {
		return nil, ErrEmptyLoadBalancer
	}
	var s []*Weighted[E]
	switch items := any(items).(type) {
	case []E:
		for _, i := range items {
			s = append(s, &Weighted[E]{i, 1})
		}
	case []*Weighted[E]:
		s = items
	}
	r := new(roundrobin[E])
	r.m = new(cache.Map[[2]int64, *Weighted[E]])
	for _, i := range s {
		r.m.Store([2]int64{r.n, r.n + i.Weight}, i)
		r.n += i.Weight
	}
	return r, nil
}

func RoundRobin[E any](items ...E) (LoadBalancer[E], error) {
	return newRoundRobin[E](items)
}

func WeightedRoundRobin[E any](items ...*Weighted[E]) (LoadBalancer[E], error) {
	return newRoundRobin[E](items)
}

func (r *roundrobin[E]) get(n int64) (e E) {
	if r.m == nil {
		panic("load balancer is closed")
	}
	r.m.Range(func(i [2]int64, w *Weighted[E]) bool {
		if n >= i[0] && n < i[1] {
			e = w.Item
			return false
		}
		return true
	})
	return
}

func (r *roundrobin[E]) Next() (next E) {
	return r.get(r.next.Swap((r.next.Load() + 1) % r.n))

}

func (r *roundrobin[E]) Close() {
	r.m = nil
}
