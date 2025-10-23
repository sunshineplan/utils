package container

import (
	"cmp"
	"sync"
	"unsafe"
)

// Element is a thread-safe element of a linked list.
type Element[T any] struct {
	mu sync.RWMutex

	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	next, prev *Element[T]

	// The list to which this element belongs.
	list *List[T]

	// The value stored with this element.
	value T
}

// Set assigns value v to the element and returns it.
func (e *Element[T]) Set(v T) *Element[T] {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.value = v
	return e
}

// Value returns the value stored with this element.
func (e *Element[T]) Value() T {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.value
}

func (e *Element[T]) nextElement() *Element[T] {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// Next returns the next list element or nil.
func (e *Element[T]) Next() *Element[T] {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.list != nil {
		e.list.mu.RLock()
		defer e.list.mu.RUnlock()
	}
	return e.nextElement()
}

func (e *Element[T]) prevElement() *Element[T] {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// Prev returns the previous list element or nil.
func (e *Element[T]) Prev() *Element[T] {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.list != nil {
		e.list.mu.RLock()
		defer e.list.mu.RUnlock()
	}
	return e.prevElement()
}

// List represents a thread-safe doubly linked list.
// The zero value for List is an empty list ready to use.
type List[T any] struct {
	mu   sync.RWMutex
	root Element[T] // sentinel list element, only &root, root.prev, and root.next are used
	len  int        // current list length excluding (this) sentinel element
}

func (l *List[T]) init() {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
}

// Init initializes or clears list l.
func (l *List[T]) Init() *List[T] {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.init()
	return l
}

// NewList returns an initialized list.
func NewList[T any]() *List[T] { return new(List[T]).Init() }

// Len returns the number of elements of list l.
// The complexity is O(1).
func (l *List[T]) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.len
}

func (l *List[T]) front() *Element[T] {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Front returns the first element of list l or nil if the list is empty.
func (l *List[T]) Front() *Element[T] {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.front()
}

func (l *List[T]) back() *Element[T] {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// Back returns the last element of list l or nil if the list is empty.
func (l *List[T]) Back() *Element[T] {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.back()
}

// lazyInit lazily initializes a zero List value.
func (l *List[T]) lazyInit() {
	if l.root.next == nil {
		l.init()
	}
}

// insert inserts e after at, increments l.len, and returns e.
func (l *List[T]) insert(e, at *Element[T]) *Element[T] {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.len++
	return e
}

// insertValue is a convenience wrapper for insert(&Element{Value: v}, at).
func (l *List[T]) insertValue(v T, at *Element[T]) *Element[T] {
	return l.insert(&Element[T]{value: v}, at)
}

// remove removes e from its list, decrements l.len
func (l *List[T]) remove(e *Element[T]) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.len--
}

// move moves e to next to at.
func (l *List[T]) move(e, at *Element[T]) {
	if e == at {
		return
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.Value.
// The element must not be nil.
func (l *List[T]) Remove(e *Element[T]) T {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.list == l {
		l.mu.Lock()
		defer l.mu.Unlock()
		// if e.list == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero Element) and l.remove will crash
		l.remove(e)
	}
	return e.value
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *List[T]) PushFront(v T) *Element[T] {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lazyInit()
	return l.insertValue(v, &l.root)
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *List[T]) PushBack(v T) *Element[T] {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lazyInit()
	return l.insertValue(v, l.root.prev)
}

// InsertBefore inserts a new element e with value v immediately before mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List[T]) InsertBefore(v T, mark *Element[T]) *Element[T] {
	mark.mu.RLock()
	defer mark.mu.RUnlock()
	if mark.list != l {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark.prev)
}

// InsertAfter inserts a new element e with value v immediately after mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List[T]) InsertAfter(v T, mark *Element[T]) *Element[T] {
	mark.mu.RLock()
	defer mark.mu.RUnlock()
	if mark.list != l {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark)
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List[T]) MoveToFront(e *Element[T]) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.list != l {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.root.next == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, &l.root)
}

// MoveToBack moves element e to the back of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List[T]) MoveToBack(e *Element[T]) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.list != l {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.root.prev == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, l.root.prev)
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List[T]) MoveBefore(e, mark *Element[T]) {
	if e == mark {
		return
	}
	unlock := lock(&e.mu, &mark.mu, true, true)
	defer unlock()
	if e.list != l || mark.list != l {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.move(e, mark.prev)
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List[T]) MoveAfter(e, mark *Element[T]) {
	if e == mark {
		return
	}
	unlock := lock(&e.mu, &mark.mu, true, true)
	defer unlock()
	if e.list != l || mark.list != l {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.move(e, mark)
}

// PushBackList inserts a copy of another list at the back of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List[T]) PushBackList(other *List[T]) {
	unlock := lock(&l.mu, &other.mu, false, true)
	defer unlock()
	l.lazyInit()
	for i, e := other.len, other.front(); i > 0; i, e = i-1, e.nextElement() {
		l.insertValue(e.value, l.root.prev)
	}
}

// PushFrontList inserts a copy of another list at the front of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List[T]) PushFrontList(other *List[T]) {
	unlock := lock(&l.mu, &other.mu, false, true)
	defer unlock()
	l.lazyInit()
	for i, e := other.len, other.back(); i > 0; i, e = i-1, e.prevElement() {
		l.insertValue(e.value, &l.root)
	}
}

func lock(s, r *sync.RWMutex, sReadOnly, rReadOnly bool) (unlock func()) {
	var sl sync.Locker = s
	var rl sync.Locker = r
	if sReadOnly {
		sl = s.RLocker()
	}
	if rReadOnly {
		rl = r.RLocker()
	}
	switch cmp.Compare(uintptr(unsafe.Pointer(s)), uintptr(unsafe.Pointer(r))) {
	case 0:
		if sReadOnly && rReadOnly {
			s.RLock()
			unlock = s.RUnlock
		} else {
			s.Lock()
			unlock = s.Unlock
		}
	case 1:
		sl.Lock()
		rl.Lock()
		unlock = func() {
			rl.Unlock()
			sl.Unlock()
		}
	case -1:
		rl.Lock()
		sl.Lock()
		unlock = func() {
			sl.Unlock()
			rl.Unlock()
		}
	}
	return
}
