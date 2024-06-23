package loadbalance

import "math/rand/v2"

var _ LoadBalancer[any] = &random[any]{}

type random[E any] struct {
	rr   *roundrobin[E]
	next chan int64
	c    chan struct{}
}

func Random[E any](items ...E) (LoadBalancer[E], error) {
	rr, err := newRoundRobin[E](items)
	if err != nil {
		return nil, err
	}
	return &random[E]{rr: rr, c: make(chan struct{})}, nil
}

func WeightedRandom[E any](items ...*Weighted[E]) (LoadBalancer[E], error) {
	rr, err := newRoundRobin[E](items)
	if err != nil {
		return nil, err
	}
	return &random[E]{rr: rr, c: make(chan struct{})}, nil
}

func (r *random[E]) init() {
	r.next = make(chan int64, r.rr.n)
	go func() {
		for {
			if _, ok := <-r.c; !ok {
				return
			}
			var s []int64
			for i := range r.rr.n {
				s = append(s, i)
			}
			rand.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
			for _, i := range s {
				r.next <- i
			}
		}
	}()
}

func (r *random[E]) Next() E {
	if r.rr == nil {
		panic("load balancer is closed")
	}
	if r.next == nil {
		r.init()
	}
	if len(r.next) <= int(r.rr.n/4) {
		r.c <- struct{}{}
	}
	return r.rr.get(<-r.next)
}

func (r *random[E]) Close() {
	if r.next != nil {
		close(r.next)
	}
	close(r.c)
	r.rr = nil
}
