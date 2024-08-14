package cache

import (
	"context"
	"log"
	"sync"
	"time"
)

var valueKey int

type item[T any] struct {
	sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	lifecycle time.Duration
	fn        func() (T, error)
}

func (i *item[T]) set(value *T) {
	if i.ctx = context.WithValue(context.Background(), &valueKey, value); i.lifecycle > 0 {
		i.ctx, i.cancel = context.WithTimeout(i.ctx, i.lifecycle)
	}
}

func (i *item[T]) value() T {
	return *i.ctx.Value(&valueKey).(*T)
}

func (i *item[T]) renew() T {
	v, err := i.fn()
	i.Lock()
	defer i.Unlock()
	if err != nil {
		log.Print(err)
		v = i.value()
	}
	i.set(&v)
	return v
}

// Cache is cache struct.
type Cache[Key, Value any] struct {
	m         Map[Key, *item[Value]]
	autoRenew bool
}

// New creates a new cache with auto clean or not.
func New[Key, Value any](autoRenew bool) *Cache[Key, Value] {
	return &Cache[Key, Value]{autoRenew: autoRenew}
}

// Set sets cache value for a key, if fn is presented, this value will regenerate when expired.
func (c *Cache[Key, Value]) Set(key Key, value Value, lifecycle time.Duration, fn func() (Value, error)) {
	i := &item[Value]{lifecycle: lifecycle, fn: fn}
	i.Lock()
	defer i.Unlock()
	i.set(&value)
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

func (c *Cache[Key, Value]) get(key Key) (i *item[Value], ok bool) {
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
func (c *Cache[Key, Value]) Get(key Key) (value Value, ok bool) {
	var i *item[Value]
	if i, ok = c.get(key); !ok {
		return
	}
	i.Lock()
	defer i.Unlock()
	value = i.value()
	return
}

// Delete deletes the value for a key.
func (c *Cache[Key, Value]) Delete(key Key) {
	if i, ok := c.m.LoadAndDelete(key); ok {
		i.Lock()
		defer i.Unlock()
		if i.cancel != nil {
			i.cancel()
		}
	}
}

// Swap swaps the value for a key and returns the previous value if any. The loaded result reports whether the key was present.
func (c *Cache[Key, Value]) Swap(key Key, value Value) (previous Value, loaded bool) {
	var i *item[Value]
	if i, loaded = c.get(key); loaded {
		i.Lock()
		defer i.Unlock()
		previous = i.value()
		i.set(&value)
	}
	return
}

// Clear deletes all values in cache.
func (c *Cache[Key, Value]) Clear() {
	c.m.Range(func(k Key, i *item[Value]) bool {
		c.m.Delete(k)
		i.Lock()
		defer i.Unlock()
		if i.cancel != nil {
			i.cancel()
		}
		return true
	})
}
