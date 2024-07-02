package cache

import (
	"container/list"
	"sync"
)

// Element is an element of a linked list.
type Element[T any] struct {
	e *list.Element

	// The list to which this element belongs.
	list *List[T]
}

// Value returns the value stored with this element.
func (e *Element[T]) Value() T {
	e.list.mu.RLock()
	defer e.list.mu.RUnlock()
	return e.e.Value.(T)
}

// Next returns the next list element or nil.
func (e *Element[T]) Next() *Element[T] {
	e.list.mu.RLock()
	defer e.list.mu.RUnlock()
	if next := e.e.Next(); next != nil {
		return &Element[T]{next, e.list}
	}
	return nil
}

// Prev returns the previous list element or nil.
func (e *Element[T]) Prev() *Element[T] {
	e.list.mu.RLock()
	defer e.list.mu.RUnlock()
	if prev := e.e.Prev(); prev != nil {
		return &Element[T]{prev, e.list}
	}
	return nil
}

// List represents a doubly linked list.
// The zero value for List is an empty list ready to use.
type List[T any] struct {
	mu sync.RWMutex
	l  list.List
}

// Init initializes or clears list l.
func (l *List[T]) Init() *List[T] {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.l.Init()
	return l
}

// New returns an initialized list.
func NewList[T any]() *List[T] { return new(List[T]).Init() }

// Len returns the number of elements of list l. The complexity is O(1).
func (l *List[T]) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.l.Len()
}

// Front returns the first element of list l or nil if the list is empty.
func (l *List[T]) Front() *Element[T] {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if e := l.l.Front(); e != nil {
		return &Element[T]{e, l}
	}
	return nil
}

// Back returns the last element of list l or nil if the list is empty.
func (l *List[T]) Back() *Element[T] {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if e := l.l.Back(); e != nil {
		return &Element[T]{e, l}
	}
	return nil
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.Value.
// The element must not be nil.
func (l *List[T]) Remove(e *Element[T]) T {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.l.Remove(e.e).(T)
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *List[T]) PushFront(v T) *Element[T] {
	l.mu.Lock()
	defer l.mu.Unlock()
	return &Element[T]{l.l.PushFront(v), l}
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *List[T]) PushBack(v T) *Element[T] {
	l.mu.Lock()
	defer l.mu.Unlock()
	return &Element[T]{l.l.PushBack(v), l}
}

// InsertBefore inserts a new element e with value v immediately before mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List[T]) InsertBefore(v T, mark *Element[T]) *Element[T] {
	l.mu.Lock()
	defer l.mu.Unlock()
	return &Element[T]{l.l.InsertBefore(v, mark.e), l}
}

// InsertAfter inserts a new element e with value v immediately after mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List[T]) InsertAfter(v T, mark *Element[T]) *Element[T] {
	l.mu.Lock()
	defer l.mu.Unlock()
	return &Element[T]{l.l.InsertAfter(v, mark.e), l}
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List[T]) MoveToFront(e *Element[T]) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.l.MoveToFront(e.e)
}

// MoveToBack moves element e to the back of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List[T]) MoveToBack(e *Element[T]) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.l.MoveToBack(e.e)
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List[T]) MoveBefore(e, mark *Element[T]) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.l.MoveBefore(e.e, mark.e)
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List[T]) MoveAfter(e, mark *Element[T]) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.l.MoveAfter(e.e, mark.e)
}

// PushBackList inserts a copy of another list at the back of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List[T]) PushBackList(other *List[T]) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.l.PushBackList(&other.l)
}

// PushFrontList inserts a copy of another list at the front of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List[T]) PushFrontList(other *List[T]) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.l.PushFrontList(&other.l)
}
