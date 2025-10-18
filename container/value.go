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

// Load returns the value set by the most recent [Value.Store].
// If there has been no call to [Value.Store] for this Value, it returns the
// zero value of T.
func (v *Value[T]) Load() (val T) {
	if loaded := v.v.Load(); loaded != nil {
		return loaded.(T)
	}
	return
}

// Store sets the value of the [Value] v to val.
func (v *Value[T]) Store(val T) {
	v.v.Store(val)
}

// Swap stores the new value into the Value and returns the previous value.
// If no value was previously stored, it returns the zero value of T.
func (v *Value[T]) Swap(new T) (old T) {
	if prev := v.v.Swap(new); prev != nil {
		return prev.(T)
	}
	return
}

// CompareAndSwap executes the compare-and-swap operation for the [Value].
func (v *Value[T]) CompareAndSwap(old, new T) (swapped bool) {
	return v.v.CompareAndSwap(old, new)
}
