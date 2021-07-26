package cache

import (
	"log"
	"sync"
	"time"
)

type item struct {
	sync.Mutex
	Value      interface{}
	Duration   time.Duration
	Expiration int64
	Regenerate func() (interface{}, error)
}

func (i *item) Expired() bool {
	if i.Expiration == 0 {
		return false
	}

	return time.Now().UnixNano() > i.Expiration
}

// Cache is cache struct.
type Cache struct {
	cache     sync.Map
	autoClean bool
}

// New creates a new cache with auto clean or not.
func New(autoClean bool) *Cache {
	c := &Cache{autoClean: autoClean}

	if autoClean {
		go c.check()
	}

	return c
}

// Set sets cache value for a key, if f is presented, this value will regenerate when expired.
func (c *Cache) Set(key, value interface{}, d time.Duration, f func() (interface{}, error)) {
	c.cache.Store(key, &item{
		Value:      value,
		Duration:   d,
		Expiration: time.Now().Add(d).UnixNano(),
		Regenerate: f,
	})
}

func (c *Cache) regenerate(i *item) {
	i.Expiration = 0
	f := i.Regenerate
	i.Unlock()

	go func() {
		value, err := f()

		i.Lock()
		defer i.Unlock()

		if err != nil {
			log.Print(err)
		} else {
			i.Value = value
		}
		i.Expiration = time.Now().Add(i.Duration).UnixNano()
	}()
}

// Get gets cache value by key and whether value was found.
func (c *Cache) Get(key interface{}) (interface{}, bool) {
	value, ok := c.cache.Load(key)
	if !ok {
		return nil, false
	}

	i := value.(*item)

	i.Lock()
	v := i.Value
	expired := i.Expired()
	f := i.Regenerate

	if expired && !c.autoClean {
		if f == nil {
			c.cache.Delete(key)
			i.Unlock()

			return nil, false
		}

		defer c.regenerate(i)

		return v, true
	}

	i.Unlock()

	return v, true
}

// Delete deletes the value for a key.
func (c *Cache) Delete(key interface{}) {
	c.cache.Delete(key)
}

// Empty deletes all values in cache.
func (c *Cache) Empty() {
	c.cache.Range(func(key, _ interface{}) bool {
		c.cache.Delete(key)
		return true
	})
}

func (c *Cache) check() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.cache.Range(func(key, value interface{}) bool {
			i := value.(*item)

			i.Lock()
			expired := i.Expired()
			f := i.Regenerate

			if expired {
				if f == nil {
					c.cache.Delete(key)
					i.Unlock()
				} else {
					defer c.regenerate(i)
				}

				return true
			}

			i.Unlock()

			return true
		})
	}
}
