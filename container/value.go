package container

import "sync/atomic"

// A Value provides an atomic load and store of a specified typed value.
// Once [Value.Store] has been called, a Value must not be copied.
//
// A Value must not be copied after first use.
type Value[T any] struct {
	v atomic.Value
}

// NewValue creates a new instance of [Value].
func NewValue[T any]() *Value[T] {
	return &Value[T]{}
}

// Load returns the value set by the most recent Store and a boolean
// indicating whether a value was stored.
// If there has been no call to Store for this Value, it returns the
// zero value of T and false.
func (v *Value[T]) Load() (val T, stored bool) {
	if v := v.v.Load(); v == nil {
		return
	} else {
		return v.(T), true
	}
}

// MustLoad returns the value set by the most recent Store.
// It panics if there has been no call to Store for this Value.
func (v *Value[T]) MustLoad() (val T) {
	if v, stored := v.Load(); stored {
		return v
	}
	panic("cache/value: there has been no call to Store for this Value")
}

// Store sets the value of the [Value] v to val.
func (v *Value[T]) Store(val T) {
	if any(val) == nil {
		panic("cache/value: store of nil value into Value")
	}
	v.v.Store(val)
}

// Swap stores the new value into the Value and returns the previous value.
// If no value was previously stored, it returns the zero value of T and false.
func (v *Value[T]) Swap(new T) (old T, stored bool) {
	if any(new) == nil {
		panic("cache/value: swap of nil value into Value")
	}
	if v := v.v.Swap(new); v == nil {
		return
	} else {
		return v.(T), true
	}
}

// CompareAndSwap executes the compare-and-swap operation for the [Value].
func (v *Value[T]) CompareAndSwap(old, new T) (swapped bool) {
	if any(new) == nil {
		panic("cache/value: compare and swap of nil value into Value")
	}
	return v.v.CompareAndSwap(old, new)
}
