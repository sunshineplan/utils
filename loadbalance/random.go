package loadbalance

import (
	"math/rand/v2"

	"github.com/sunshineplan/utils/container"
)

var _ LoadBalancer[any] = &random[any]{}

// random implements a thread-safe random load balancer by extending the round-robin load balancer.
// It selects elements randomly from the ring, ensuring thread-safe operations using the
// underlying round-robin mutex and ring mutexes.
type random[E any] struct {
	*roundrobin[E] // Embeds roundrobin for shared functionality.
}

func newRandom[E any, Items []E | []Weighted[E]](items Items) (*random[E], error) {
	lb, err := newRoundRobin[E](items)
	if err != nil {
		return nil, err
	}
	return &random[E]{lb}, nil
}

// Random creates a new random load balancer with the given items.
// Each item appears once in the pool. It returns error with ErrEmptyLoadBalancer if no items are provided.
func Random[E any](items ...E) (LoadBalancer[E], error) {
	return newRandom[E](items)
}

// WeightedRandom creates a new weighted random load balancer.
// Each item's weight determines how many times it appears in the pool.
// It returns error with ErrEmptyLoadBalancer if no items have positive weight.
func WeightedRandom[E any](items ...Weighted[E]) (LoadBalancer[E], error) {
	return newRandom[E](items)
}

// RandomFromRing creates a new random load balancer from an existing ring.
// It uses the provided ring directly, ensuring thread-safe random selection.
// It returns error with ErrEmptyLoadBalancer if the ring is nil or empty.
func RandomFromRing[E any](ring *container.Ring[E]) (LoadBalancer[E], error) {
	len := ring.Len()
	if len == 0 {
		return nil, ErrEmptyLoadBalancer
	}
	return &random[E]{&roundrobin[E]{ring: ring, len: len}}, nil
}

// Next returns a randomly selected element from the load balancer.
// It is thread-safe, using a write lock to update the ring position.
// If the balancer is empty, it returns the zero value of E.
func (r *random[E]) Next() E {
	r.Lock()
	defer r.Unlock()
	r.ring = r.ring.Move(rand.IntN(r.len))
	return r.ring.Value()
}
