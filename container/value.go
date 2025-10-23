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

// An Int provides atomic operations on an integer value of type T.
// It is a generic wrapper around [atomic.Int64], allowing usage with
// signed integer types such as int, int8, int16, int32, and int64.
type Int[T ~int64 | ~int32 | ~int16 | ~int8 | ~int] struct {
	v atomic.Int64
}

// Add atomically adds delta to the value and returns the new value.
func (v *Int[T]) Add(delta T) (new T) {
	return T(v.v.Add(int64(delta)))
}

// And atomically performs a bitwise AND operation with mask and returns
// the previous value.
func (v *Int[T]) And(mask T) (old T) {
	return T(v.v.And(int64(mask)))
}

// CompareAndSwap executes the compare-and-swap operation for the [Int].
// It compares the current value with old, and if they are equal,
// sets it to new and returns true. Otherwise, it returns false.
func (v *Int[T]) CompareAndSwap(old T, new T) (swapped bool) {
	return v.v.CompareAndSwap(int64(old), int64(new))
}

// Load atomically loads and returns the current value.
func (v *Int[T]) Load() T {
	return T(v.v.Load())
}

// Or atomically performs a bitwise OR operation with mask and returns
// the previous value.
func (v *Int[T]) Or(mask T) (old T) {
	return T(v.v.Or(int64(mask)))
}

// Store atomically stores val into the [Int].
func (v *Int[T]) Store(val T) {
	v.v.Store(int64(val))
}

// Swap atomically stores new into the [Int] and returns the previous value.
func (v *Int[T]) Swap(new T) (old T) {
	return T(v.v.Swap(int64(new)))
}
