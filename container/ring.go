package container

import (
	"container/ring"
	"sync"
)

var ringMutex sync.RWMutex

type mutex ring.Ring

func newMutex() *mutex {
	r := ring.New(1)
	r.Value = new(sync.RWMutex)
	return (*mutex)(r)
}

func (mu *mutex) Lock() {
	ringMutex.RLock()
	defer ringMutex.RUnlock()
	(*ring.Ring)(mu).Do(func(a any) {
		a.(*sync.RWMutex).Lock()
	})
}

func (mu *mutex) Unlock() {
	ringMutex.RLock()
	defer ringMutex.RUnlock()
	(*ring.Ring)(mu).Do(func(a any) {
		a.(*sync.RWMutex).Unlock()
	})
}

func (mu *mutex) RLock() {
	ringMutex.RLock()
	defer ringMutex.RUnlock()
	(*ring.Ring)(mu).Do(func(a any) {
		a.(*sync.RWMutex).RLock()
	})
}

func (mu *mutex) RUnlock() {
	ringMutex.RLock()
	defer ringMutex.RUnlock()
	(*ring.Ring)(mu).Do(func(a any) {
		a.(*sync.RWMutex).RUnlock()
	})
}

func (mu *mutex) Link(s *mutex) *mutex {
	ringMutex.Lock()
	defer ringMutex.Unlock()
	return (*mutex)((*ring.Ring)(mu).Link((*ring.Ring)(s)))
}

// A Ring is an element of a circular list, or ring.
// Rings do not have a beginning or end; a pointer to any ring element
// serves as reference to the entire ring. Empty rings are represented
// as nil Ring pointers. The zero value for a Ring is a one-element
// ring with a nil Value.
type Ring[T any] struct {
	mu *mutex
	r  *ring.Ring
}

// Next returns the next ring element. r must not be empty.
func (r *Ring[T]) Next() *Ring[T] {
	r.mu.Lock()
	defer r.mu.Unlock()
	return &Ring[T]{r.mu, r.r.Next()}
}

// Prev returns the previous ring element. r must not be empty.
func (r *Ring[T]) Prev() *Ring[T] {
	r.mu.Lock()
	defer r.mu.Unlock()
	return &Ring[T]{r.mu, r.r.Prev()}
}

// Move moves n % r.Len() elements backward (n < 0) or forward (n >= 0)
// in the ring and returns that ring element. r must not be empty.
func (r *Ring[T]) Move(n int) *Ring[T] {
	r.mu.Lock()
	defer r.mu.Unlock()
	return &Ring[T]{newMutex(), r.r.Move(n)}
}

// NewRing creates a ring of n elements.
func NewRing[T any](n int) *Ring[T] {
	if n <= 0 {
		return nil
	}
	return &Ring[T]{newMutex(), ring.New(n)}
}

func (r *Ring[T]) Set(v T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.r.Value = &v
}

func (r *Ring[T]) Value() *T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, _ := r.r.Value.(*T)
	return v
}

// Link connects ring r with ring s such that r.Next()
// becomes s and returns the original value for r.Next().
// r must not be empty.
//
// If r and s point to the same ring, linking
// them removes the elements between r and s from the ring.
// The removed elements form a subring and the result is a
// reference to that subring (if no elements were removed,
// the result is still the original value for r.Next(),
// and not nil).
//
// If r and s point to different rings, linking
// them creates a single ring with the elements of s inserted
// after r. The result points to the element following the
// last element of s after insertion.
func (r *Ring[T]) Link(s *Ring[T]) *Ring[T] {
	if s == nil {
		return r.Next()
	}
	m := r.mu.Link(s.mu)
	r.mu.Lock()
	defer r.mu.Unlock()
	return &Ring[T]{m, r.r.Link(s.r)}
}

// Unlink removes n % r.Len() elements from the ring r, starting
// at r.Next(). If n % r.Len() == 0, r remains unchanged.
// The result is the removed subring. r must not be empty.
func (r *Ring[T]) Unlink(n int) *Ring[T] {
	r.mu.Lock()
	defer r.mu.Unlock()
	u := r.r.Unlink(n)
	if u == nil {
		return nil
	}
	return &Ring[T]{newMutex(), u}
}

// Len computes the number of elements in ring r.
// It executes in time proportional to the number of elements.
func (r *Ring[T]) Len() int {
	if r == nil {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.r.Len()
}

// Do calls function f on each element of the ring, in forward order.
// The behavior of Do is undefined if f changes *r.
func (r *Ring[T]) Do(f func(*T)) {
	if r == nil {
		return
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	r.r.Do(func(a any) {
		if v, ok := a.(*T); ok {
			f(v)
		} else {
			f(nil)
		}
	})
}
