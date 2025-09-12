package cache

import (
	"runtime"
	"weak"

	"github.com/sunshineplan/utils/container"
)

// Cache is cache struct.
type Cache[Key any, Value any] struct {
	m container.Map[weak.Pointer[Key], Value]
}

// New creates a new cache with auto clean or not.
func New[Key any, Value any]() *Cache[Key, Value] {
	return &Cache[Key, Value]{}
}

// Set sets cache value for a key.
func (c *Cache[Key, Value]) Set(key *Key, value Value) {
	p := weak.Make(key)
	c.m.Store(p, value)
	runtime.AddCleanup(key, func(p weak.Pointer[Key]) { c.m.Delete(p) }, p)
}

// Get gets cache value by key and whether value was found.
func (c *Cache[Key, Value]) Get(key *Key) (value Value, ok bool) {
	p := weak.Make(key)
	return c.m.Load(p)
}

// Delete deletes the value for a key.
func (c *Cache[Key, Value]) Delete(key *Key) {
	p := weak.Make(key)
	c.m.Delete(p)
}

// Swap swaps the value for a key and returns the previous value if any. The loaded result reports whether the key was present.
func (c *Cache[Key, Value]) Swap(key *Key, value Value) (previous Value, loaded bool) {
	p := weak.Make(key)
	return c.m.Swap(p, value)
}

// Clear deletes all values in cache.
func (c *Cache[Key, Value]) Clear() {
	c.m.Clear()
}
