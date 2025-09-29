package loadbalance

import (
	"sync"

	"github.com/sunshineplan/utils/container"
)

var _ LoadBalancer[any] = &roundrobin[any]{}

// roundrobin implements a thread-safe round-robin load balancer using a circular ring.
// It supports both simple and weighted item distributions, ensuring fair rotation through
// elements. All methods are thread-safe using a struct-level mutex and the underlying ring's mutexes.
type roundrobin[E any] struct {
	sync.RWMutex
	ring *container.Ring[E] // The underlying ring storing the elements.
	len  int                // Cached length of the ring.
}

func newRoundRobin[E any, Items []E | []Weighted[E]](items Items) (*roundrobin[E], error) {
	if len(items) == 0 {
		return nil, ErrEmptyLoadBalancer
	}
	var ring *container.Ring[E]
	var n int
	switch items := any(items).(type) {
	case []E:
		n = len(items)
		ring = container.NewRing[E](n)
		for _, i := range items {
			ring = ring.Set(i).Next()
		}
	case []Weighted[E]:
		for _, i := range items {
			if i.Weight == 0 {
				continue
			}
			subring := container.NewRing[E](i.Weight)
			for range i.Weight {
				subring = subring.Set(i.Item).Next()
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
		if n = ring.Len(); n == 0 {
			return nil, ErrEmptyLoadBalancer
		}
	}
	return &roundrobin[E]{ring: ring, len: n}, nil
}

// RoundRobin creates a new round-robin load balancer with the given items.
// Each item appears once in the rotation. It returns error with ErrEmptyLoadBalancer if no items are provided.
func RoundRobin[E any](items ...E) (LoadBalancer[E], error) {
	return newRoundRobin[E](items)
}

// WeightedRoundRobin creates a new weighted round-robin load balancer.
// Each item's weight determines how many times it appears in the rotation.
// It returns error with ErrEmptyLoadBalancer if no items have positive weight.
func WeightedRoundRobin[E any](items ...Weighted[E]) (LoadBalancer[E], error) {
	return newRoundRobin[E](items)
}

// RoundRobinFromRing creates a new round-robin load balancer from an existing ring.
// It uses the provided ring directly, ensuring thread-safe operations.
// It returns error with ErrEmptyLoadBalancer if the ring is nil or empty.
func RoundRobinFromRing[E any](ring *container.Ring[E]) (LoadBalancer[E], error) {
	len := ring.Len()
	if len == 0 {
		return nil, ErrEmptyLoadBalancer
	}
	return &roundrobin[E]{ring: ring, len: len}, nil
}

// Len returns the number of elements in the load balancer.
// It is thread-safe and uses the cached length for O(1) access.
func (r *roundrobin[E]) Len() int {
	r.RLock()
	defer r.RUnlock()
	return r.len
}

// Next returns the next element in the round-robin sequence.
// It is thread-safe, advancing the ring to the next position.
// If the balancer is empty, it returns the zero value of E.
func (r *roundrobin[E]) Next() (next E) {
	r.Lock()
	defer r.Unlock()
	next = r.ring.Value()
	r.ring = r.ring.Next()
	return
}

// Link merges the given ring into the load balancer, inserting its elements
// after the current position. It is thread-safe, updates the cached length,
// and returns the load balancer for chaining.
func (r *roundrobin[E]) Link(s *container.Ring[E]) LoadBalancer[E] {
	r.Lock()
	defer r.Unlock()
	r.ring = r.ring.Prev().Link(s)
	r.len = r.ring.Len()
	return r
}

// Unlink removes n elements starting from the next position and returns
// the load balancer for chaining. It is thread-safe, updates the cached length,
// and sets the ring to nil if it becomes empty.
func (r *roundrobin[E]) Unlink(n int) LoadBalancer[E] {
	r.Lock()
	defer r.Unlock()
	r.ring = r.ring.Unlink(n)
	r.len = r.ring.Len()
	return r
}
