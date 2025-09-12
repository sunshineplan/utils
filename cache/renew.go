package cache

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/sunshineplan/utils/container"
)

type item[T any] struct {
	sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	lifecycle time.Duration
	value     T
	fn        func() (T, error)
}

func (i *item[T]) set(value T) {
	i.ctx = context.Background()
	if i.lifecycle > 0 {
		i.ctx, i.cancel = context.WithTimeout(i.ctx, i.lifecycle)
	}
	i.value = value
}

func (i *item[T]) renew() T {
	v, err := i.fn()
	i.Lock()
	defer i.Unlock()
	if err != nil {
		log.Print(err)
		v = i.value
	}
	i.set(v)
	return v
}

// CacheWithRenew is cache struct.
type CacheWithRenew[Key comparable, Value any] struct {
	m         container.Map[Key, *item[Value]]
	autoRenew bool
}

// NewWithRenew creates a new cache with auto clean or not.
func NewWithRenew[Key comparable, Value any](autoRenew bool) *CacheWithRenew[Key, Value] {
	return &CacheWithRenew[Key, Value]{autoRenew: autoRenew}
}

// Set sets cache value for a key, if fn is presented, this value will regenerate when expired.
func (c *CacheWithRenew[Key, Value]) Set(key Key, value Value, lifecycle time.Duration, fn func() (Value, error)) {
	i := &item[Value]{lifecycle: lifecycle, fn: fn}
	i.set(value)
	if c.autoRenew && lifecycle > 0 {
		go func() {
			for {
				<-i.ctx.Done()
				if i.ctx.Err() == context.DeadlineExceeded {
					if i.fn != nil {
						i.renew()
						continue
					} else {
						c.Delete(key)
					}
				}
				return
			}
		}()
	}
	c.m.Store(key, i)
}

func (c *CacheWithRenew[Key, Value]) get(key Key) (i *item[Value], ok bool) {
	if i, ok = c.m.Load(key); !ok {
		return
	}
	if !c.autoRenew && i.ctx.Err() == context.DeadlineExceeded {
		if i.fn == nil {
			c.Delete(key)
			return nil, false
		}
		i.renew()
	}
	return
}

// Get gets cache value by key and whether value was found.
func (c *CacheWithRenew[Key, Value]) Get(key Key) (value Value, ok bool) {
	var i *item[Value]
	if i, ok = c.get(key); !ok {
		return
	}
	i.Lock()
	defer i.Unlock()
	value = i.value
	return
}

// Delete deletes the value for a key.
func (c *CacheWithRenew[Key, Value]) Delete(key Key) {
	if i, ok := c.m.LoadAndDelete(key); ok {
		i.Lock()
		defer i.Unlock()
		if i.cancel != nil {
			i.cancel()
		}
	}
}

// Swap swaps the value for a key and returns the previous value if any. The loaded result reports whether the key was present.
func (c *CacheWithRenew[Key, Value]) Swap(key Key, value Value) (previous Value, loaded bool) {
	var i *item[Value]
	if i, loaded = c.get(key); loaded {
		i.Lock()
		defer i.Unlock()
		previous = i.value
		i.set(value)
	}
	return
}

// Clear deletes all values in cache.
func (c *CacheWithRenew[Key, Value]) Clear() {
	c.m.Range(func(key Key, i *item[Value]) bool {
		c.m.Delete(key)
		i.Lock()
		defer i.Unlock()
		if i.cancel != nil {
			i.cancel()
		}
		return true
	})
}
