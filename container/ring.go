package container

import (
	"sync"
	"unsafe"
)

// A Ring is an element of a circular list, or ring.
// Rings do not have a beginning or end; a pointer to any ring element
// serves as reference to the entire ring. Empty rings are represented
// as nil Ring pointers. The zero value for a Ring is a one-element
// ring with a nil Value.
type Ring[T any] struct {
	mu     sync.RWMutex
	ringMu *sync.RWMutex

	next, prev *Ring[T]
	value      T // for use by client; untouched by this library
}

func (r *Ring[T]) init() *Ring[T] {
	r.ringMu = new(sync.RWMutex)
	r.next = r
	r.prev = r
	return r
}

// Next returns the next ring element. r must not be empty.
func (r *Ring[T]) Next() *Ring[T] {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ringMu == nil {
		return r.init()
	}
	r.ringMu.RLock()
	defer r.ringMu.RUnlock()
	return r.next
}

// Prev returns the previous ring element. r must not be empty.
func (r *Ring[T]) Prev() *Ring[T] {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ringMu == nil {
		return r.init()
	}
	r.ringMu.RLock()
	defer r.ringMu.RUnlock()
	return r.prev
}

func (r *Ring[T]) move(n int) *Ring[T] {
	switch {
	case n < 0:
		for ; n < 0; n++ {
			r = r.prev
		}
	case n > 0:
		for ; n > 0; n-- {
			r = r.next
		}
	}
	return r
}

// Move moves n % r.Len() elements backward (n < 0) or forward (n >= 0)
// in the ring and returns that ring element. r must not be empty.
func (r *Ring[T]) Move(n int) *Ring[T] {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ringMu == nil {
		return r.init()
	}
	r.ringMu.RLock()
	defer r.ringMu.RUnlock()
	return r.move(n)
}

// NewRing creates a ring of n elements.
func NewRing[T any](n int) *Ring[T] {
	if n <= 0 {
		return nil
	}
	r := &Ring[T]{ringMu: new(sync.RWMutex)}
	p := r
	for i := 1; i < n; i++ {
		p.next = &Ring[T]{ringMu: p.ringMu, prev: p}
		p = p.next
	}
	p.next = r
	r.prev = p
	return r
}

func linkLock[T any](r, s *Ring[T]) (unlock func()) {
	rmu, smu := r.ringMu, s.ringMu
	if s == r {
		r.mu.Lock()
		rmu.Lock()
		unlock = func() {
			rmu.Unlock()
			r.mu.Unlock()
		}
	} else {
		order := uintptr(unsafe.Pointer(&r.mu)) < uintptr(unsafe.Pointer(&s.mu))
		var finalUnlock func()
		if order {
			r.mu.Lock()
			s.mu.Lock()
			finalUnlock = func() {
				s.mu.Unlock()
				r.mu.Unlock()
			}
		} else {
			s.mu.Lock()
			r.mu.Lock()
			finalUnlock = func() {
				r.mu.Unlock()
				s.mu.Unlock()
			}
		}
		switch smu {
		case nil:
			s.init()
			smu = rmu
			fallthrough
		case rmu:
			smu.Lock()
			unlock = func() {
				smu.Unlock()
				finalUnlock()
			}
		default:
			if order {
				rmu.Lock()
				smu.Lock()
				unlock = func() {
					smu.Unlock()
					rmu.Unlock()
					finalUnlock()
				}
			} else {
				smu.Lock()
				rmu.Lock()
				unlock = func() {
					rmu.Unlock()
					smu.Unlock()
					finalUnlock()
				}
			}
		}
	}
	return
}

func (r *Ring[T]) link(s *Ring[T]) *Ring[T] {
	var sameRing bool
	if s.ringMu == r.ringMu {
		sameRing = true
	} else {
		s.ringMu = r.ringMu
		for p := s.next; p != s; p = p.next {
			p.ringMu = r.ringMu
		}
	}
	n := r.next
	p := s.prev
	// Note: Cannot use multiple assignment because
	// evaluation order of LHS is not specified.
	r.next = s
	s.prev = r
	n.prev = p
	p.next = n
	if sameRing {
		n.ringMu = new(sync.RWMutex)
		for p := n.next; p != n; p = p.next {
			p.ringMu = n.ringMu
		}
	}
	return n
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
	r.mu.Lock()
	if r.ringMu == nil {
		r.init()
	}
	n := r.next
	r.mu.Unlock()
	if s == nil {
		return n
	}
	unlock := linkLock(r, s)
	defer unlock()
	return r.link(s)
}

// Unlink removes n % r.Len() elements from the ring r, starting
// at r.Next(). If n % r.Len() == 0, r remains unchanged.
// The result is the removed subring. r must not be empty.
func (r *Ring[T]) Unlink(n int) *Ring[T] {
	if n <= 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ringMu == nil {
		return r.init()
	}
	r.ringMu.Lock()
	defer r.ringMu.Unlock()
	return r.link(r.move(n + 1))
}

// Len computes the number of elements in ring r.
// It executes in time proportional to the number of elements.
func (r *Ring[T]) Len() int {
	n := 0
	if r != nil {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.ringMu == nil {
			r.init()
			return 1
		}
		r.ringMu.RLock()
		defer r.ringMu.RUnlock()
		n = 1
		for p := r.next; p != r; p = p.next {
			n++
		}
	}
	return n
}

// Do calls function f on each element of the ring, in forward order.
// The behavior of Do is undefined if f changes *r.
func (r *Ring[T]) Do(f func(T)) {
	if r != nil {
		r.mu.RLock()
		if r.ringMu == nil {
			r.mu.RUnlock()
			return
		}
		r.ringMu.RLock()
		defer r.ringMu.RUnlock()
		f(r.value)
		r.mu.RUnlock()
		for p := r.next; p != r; p = p.next {
			p.mu.RLock()
			f(p.value)
			p.mu.RUnlock()
		}
	}
}

func (r *Ring[T]) Set(v T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ringMu == nil {
		r.init()
	}
	r.value = v
}

func (r *Ring[T]) Value() T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.value
}
