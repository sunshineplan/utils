package container

import "sync"

// Map is a generic concurrency-safe map that wraps sync.Map
// and provides type-safe access for keys and values.
type Map[Key comparable, Value any] struct {
	m sync.Map
}

// NewMap creates and returns a new, empty generic concurrency-safe Map.
func NewMap[Key comparable, Value any]() *Map[Key, Value] {
	return &Map[Key, Value]{}
}

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *Map[Key, Value]) Load(key Key) (value Value, ok bool) {
	var v any
	if v, ok = m.m.Load(key); ok {
		value, _ = v.(Value)
	}
	return
}

// Store sets the value for a key.
func (m *Map[Key, Value]) Store(key Key, value Value) {
	m.m.Store(key, value)
}

// Clear deletes all the entries, resulting in an empty Map.
func (m *Map[Key, Value]) Clear() {
	m.m.Clear()
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *Map[Key, Value]) LoadOrStore(key Key, value Value) (actual Value, loaded bool) {
	var v any
	if v, loaded = m.m.LoadOrStore(key, value); loaded {
		actual, _ = v.(Value)
	} else {
		actual = value
	}
	return
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (m *Map[Key, Value]) LoadAndDelete(key Key) (value Value, loaded bool) {
	var v any
	if v, loaded = m.m.LoadAndDelete(key); loaded {
		value, _ = v.(Value)
	}
	return
}

// Delete deletes the value for a key.
func (m *Map[Key, Value]) Delete(key Key) {
	m.m.Delete(key)
}

// Swap swaps the value for a key and returns the previous value if any.
// The loaded result reports whether the key was present.
func (m *Map[Key, Value]) Swap(key Key, value Value) (previous Value, loaded bool) {
	var v any
	if v, loaded = m.m.Swap(key, value); loaded {
		previous, _ = v.(Value)
	}
	return
}

// CompareAndSwap swaps the old and new values for key
// if the value stored in the map is equal to old.
// The old value must be of a comparable type.
func (m *Map[Key, Value]) CompareAndSwap(key Key, old Value, new Value) (swapped bool) {
	return m.m.CompareAndSwap(key, old, new)
}

// CompareAndDelete deletes the entry for key if its value is equal to old.
// The old value must be of a comparable type.
//
// If there is no current value for key in the map, CompareAndDelete
// returns false (even if the old value is the nil interface value).
func (m *Map[Key, Value]) CompareAndDelete(key Key, old Value) (deleted bool) {
	return m.m.CompareAndDelete(key, old)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently (including by f), Range may reflect any
// mapping for that key from any point during the Range call. Range does not
// block other methods on the receiver; even f itself may call any method on m.
//
// Range may be O(N) with the number of elements in the map even if f returns
// false after a constant number of calls.
func (m *Map[Key, Value]) Range(f func(Key, Value) bool) {
	m.m.Range(func(key, value any) bool {
		if k, ok := key.(Key); ok {
			if v, ok := value.(Value); ok {
				return f(k, v)
			}
		}
		return true
	})
}
