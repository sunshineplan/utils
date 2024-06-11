package cache

import (
	"context"
	"log"
	"sync"
	"time"
)

var valueKey int

func newContext(value any, lifecycle time.Duration) (ctx context.Context, cancel context.CancelFunc) {
	ctx = context.WithValue(context.Background(), &valueKey, value)
	if lifecycle > 0 {
		ctx, cancel = context.WithTimeout(ctx, lifecycle)
	}
	return
}

type item[T any] struct {
	sync.Mutex
	context.Context
	cancel    context.CancelFunc
	lifecycle time.Duration
	fn        func() (T, error)
}

func (i *item[T]) value() T {
	i.Lock()
	defer i.Unlock()
	return i.Value(&valueKey).(T)
}

func (i *item[T]) renew() T {
	v, err := i.fn()
	if err != nil {
		log.Print(err)
		v = i.value()
	}
	i.Lock()
	defer i.Unlock()
	i.Context, i.cancel = newContext(v, i.lifecycle)
	return v
}

// Cache is cache struct.
type Cache[Key, Value any] struct {
	cache     sync.Map
	autoRenew bool
}

// New creates a new cache with auto clean or not.
func New[Key, Value any](autoRenew bool) *Cache[Key, Value] {
	return &Cache[Key, Value]{autoRenew: autoRenew}
}

// Set sets cache value for a key, if fn is presented, this value will regenerate when expired.
func (c *Cache[Key, Value]) Set(key Key, value Value, lifecycle time.Duration, fn func() (Value, error)) {
	i := &item[Value]{lifecycle: lifecycle, fn: fn}
	i.Context, i.cancel = newContext(value, lifecycle)
	if c.autoRenew && lifecycle > 0 {
		go func() {
			for {
				<-i.Done()
				if err := i.Err(); err == context.DeadlineExceeded {
					if i.fn != nil {
						i.renew()
					} else {
						c.Delete(key)
					}
				} else {
					return
				}
			}
		}()
	}
	c.cache.Store(key, i)
}

// Get gets cache value by key and whether value was found.
func (c *Cache[Key, Value]) Get(key Key) (Value, bool) {
	v, ok := c.cache.Load(key)
	if !ok {
		return *new(Value), false
	}
	if i := v.(*item[Value]); !c.autoRenew && i.Err() == context.DeadlineExceeded {
		if i.fn == nil {
			c.Delete(key)
			return *new(Value), false
		}
		return i.renew(), true
	} else {
		return i.value(), true
	}
}

// Delete deletes the value for a key.
func (c *Cache[Key, Value]) Delete(key Key) {
	if v, ok := c.cache.LoadAndDelete(key); ok {
		if v, ok := v.(*item[Value]); ok {
			if v.cancel != nil {
				v.cancel()
			}
		}
	}
}

// Empty deletes all values in cache.
func (c *Cache[Key, Value]) Empty() {
	c.cache.Range(func(k, v any) bool {
		c.cache.Delete(k)
		if v, ok := v.(*item[Value]); ok {
			if v.cancel != nil {
				v.cancel()
			}
		}
		return true
	})
}
