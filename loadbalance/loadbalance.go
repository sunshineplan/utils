package loadbalance

import (
	"errors"

	"github.com/sunshineplan/utils/container"
)

var ErrEmptyLoadBalancer = errors.New("empty load balancer")

type LoadBalancer[E any] interface {
	Len() int
	Next() E
	Link(*container.Ring[E]) LoadBalancer[E]
	Unlink(int) LoadBalancer[E]
}

type Weighted[E any] struct {
	Item   E
	Weight int
}
