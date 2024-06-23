package loadbalance

import "errors"

var ErrEmptyLoadBalancer = errors.New("empty load balancer")

type LoadBalancer[E any] interface {
	Next() E
	Close()
}

type Weighted[E any] struct {
	Item   E
	Weight int64
}
