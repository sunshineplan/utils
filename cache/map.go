package cache

import "sync"

type Map[Key, Value any] struct {
	m sync.Map
}

func NewMap[Key, Value any]() *Map[Key, Value] {
	return &Map[Key, Value]{}
}

func (m *Map[Key, Value]) Load(key Key) (value Value, ok bool) {
	var v any
	if v, ok = m.m.Load(key); ok && v != nil {
		value = v.(Value)
	}
	return
}

func (m *Map[Key, Value]) Store(key Key, value Value) {
	m.m.Store(key, value)
}

func (m *Map[Key, Value]) LoadOrStore(key Key, value Value) (actual Value, loaded bool) {
	var v any
	if v, loaded = m.m.LoadOrStore(key, value); v != nil {
		actual = v.(Value)
	}
	return
}

func (m *Map[Key, Value]) LoadAndDelete(key Key) (value Value, loaded bool) {
	var v any
	if v, loaded = m.m.LoadAndDelete(key); loaded && v != nil {
		value = v.(Value)
	}
	return
}

func (m *Map[Key, Value]) Delete(key Key) {
	m.m.Delete(key)
}

func (m *Map[Key, Value]) Swap(key Key, value Value) (previous Value, loaded bool) {
	var v any
	if v, loaded = m.m.Swap(key, value); loaded && v != nil {
		previous = v.(Value)
	}
	return
}

func (m *Map[Key, Value]) CompareAndSwap(key Key, old Value, new Value) bool {
	return m.m.CompareAndSwap(key, old, new)
}

func (m *Map[Key, Value]) CompareAndDelete(key Key, old Value) (deleted bool) {
	return m.m.CompareAndDelete(key, old)
}

func (m *Map[Key, Value]) Range(f func(Key, Value) bool) {
	m.m.Range(func(key, value any) bool {
		var k Key
		if key != nil {
			k = key.(Key)
		}
		var v Value
		if value != nil {
			v = value.(Value)
		}
		return f(k, v)
	})
}
