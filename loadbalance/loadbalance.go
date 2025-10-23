package loadbalance

import (
	"errors"

	"github.com/sunshineplan/utils/container"
)

// ErrEmptyLoadBalancer is returned when a load balancer is initialized with no valid items.
var ErrEmptyLoadBalancer = errors.New("empty load balancer")

// LoadBalancer defines an interface for load balancing algorithms.
// It provides methods to access and manipulate a circular list of elements of type E,
// ensuring thread-safe operations for concurrent use.
type LoadBalancer[E any] interface {
	// Len returns the number of elements in the load balancer.
	Len() int
	// Next returns the next element according to the load balancing strategy.
	Next() E
	// Link merges the given ring into the load balancer, inserting its elements.
	Link(*container.Ring[E]) LoadBalancer[E]
	// Unlink removes n elements from the load balancer, starting from the next position.
	Unlink(int) LoadBalancer[E]
}

// Weighted represents an item with an associated weight for weighted load balancing.
// The Weight field determines how many times the Item appears in the load balancer.
type Weighted[E any] struct {
	Item   E   // The item to be balanced.
	Weight int // The weight determining the item's frequency in the rotation.
}
