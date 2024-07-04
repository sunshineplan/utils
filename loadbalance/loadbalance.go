package loadbalance

import (
	"errors"
	"sync"

	"github.com/sunshineplan/utils/cache"
)

var ErrEmptyLoadBalancer = errors.New("empty load balancer")

var mu sync.RWMutex

type LoadBalancer[E any] interface {
	Len() int
	Next() E
	Ring() *cache.Ring[*E]
	Link(LoadBalancer[E]) LoadBalancer[E]
	Unlink(int) LoadBalancer[E]
}

type Weighted[E any] struct {
	Item   E
	Weight int
}
