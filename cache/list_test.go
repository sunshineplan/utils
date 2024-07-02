// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"container/list"
	"testing"
)

func checkListLen[T int](t *testing.T, l *List[T], len int) bool {
	if n := l.Len(); n != len {
		t.Errorf("l.Len() = %d, want %d", n, len)
		return false
	}
	return true
}

func checkList[T int](t *testing.T, l *List[T], es []T) {
	if !checkListLen(t, l, len(es)) {
		return
	}

	i := 0
	for e := l.Front(); e != nil; e = e.Next() {
		if v := e.Value(); v != es[i] {
			t.Errorf("elt[%d].Value = %v, want %v", i, v, es[i])
		}
		i++
	}
}

func TestExtending(t *testing.T) {
	l1 := NewList[int]()
	l2 := NewList[int]()

	l1.PushBack(1)
	l1.PushBack(2)
	l1.PushBack(3)

	l2.PushBack(4)
	l2.PushBack(5)

	l3 := NewList[int]()
	l3.PushBackList(l1)
	checkList(t, l3, []int{1, 2, 3})
	l3.PushBackList(l2)
	checkList(t, l3, []int{1, 2, 3, 4, 5})

	l3 = NewList[int]()
	l3.PushFrontList(l2)
	checkList(t, l3, []int{4, 5})
	l3.PushFrontList(l1)
	checkList(t, l3, []int{1, 2, 3, 4, 5})

	checkList(t, l1, []int{1, 2, 3})
	checkList(t, l2, []int{4, 5})

	l3 = NewList[int]()
	l3.PushBackList(l1)
	checkList(t, l3, []int{1, 2, 3})
	l3.PushBackList(l3)
	checkList(t, l3, []int{1, 2, 3, 1, 2, 3})

	l3 = NewList[int]()
	l3.PushFrontList(l1)
	checkList(t, l3, []int{1, 2, 3})
	l3.PushFrontList(l3)
	checkList(t, l3, []int{1, 2, 3, 1, 2, 3})

	l3 = NewList[int]()
	l1.PushBackList(l3)
	checkList(t, l1, []int{1, 2, 3})
	l1.PushFrontList(l3)
	checkList(t, l1, []int{1, 2, 3})
}

func TestIssue4103(t *testing.T) {
	l1 := NewList[int]()
	l1.PushBack(1)
	l1.PushBack(2)

	l2 := NewList[int]()
	l2.PushBack(3)
	l2.PushBack(4)

	e := l1.Front()
	l2.Remove(e) // l2 should not change because e is not an element of l2
	if n := l2.Len(); n != 2 {
		t.Errorf("l2.Len() = %d, want 2", n)
	}

	l1.InsertBefore(8, e)
	if n := l1.Len(); n != 3 {
		t.Errorf("l1.Len() = %d, want 3", n)
	}
}

func TestIssue6349(t *testing.T) {
	l := NewList[int]()
	l.PushBack(1)
	l.PushBack(2)

	e := l.Front()
	l.Remove(e)
	if v := e.Value(); v != 1 {
		t.Errorf("e.value = %d, want 1", v)
	}
	if e.Next() != nil {
		t.Errorf("e.Next() != nil")
	}
	if e.Prev() != nil {
		t.Errorf("e.Prev() != nil")
	}
}

// Test PushFront, PushBack, PushFrontList, PushBackList with uninitialized List
func TestZeroList(t *testing.T) {
	var l1 = new(List[int])
	l1.PushFront(1)
	checkList(t, l1, []int{1})

	var l2 = new(List[int])
	l2.PushBack(1)
	checkList(t, l2, []int{1})

	var l3 = new(List[int])
	l3.PushFrontList(l1)
	checkList(t, l3, []int{1})

	var l4 = new(List[int])
	l4.PushBackList(l2)
	checkList(t, l4, []int{1})
}

// Test that a list l is not modified when calling InsertBefore with a mark that is not an element of l.
func TestInsertBeforeUnknownMark(t *testing.T) {
	var l List[int]
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)
	l.InsertBefore(1, &Element[int]{new(list.Element), nil})
	checkList(t, &l, []int{1, 2, 3})
}

// Test that a list l is not modified when calling InsertAfter with a mark that is not an element of l.
func TestInsertAfterUnknownMark(t *testing.T) {
	var l List[int]
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)
	l.InsertAfter(1, &Element[int]{new(list.Element), nil})
	checkList(t, &l, []int{1, 2, 3})
}

// Test that a list l is not modified when calling MoveAfter or MoveBefore with a mark that is not an element of l.
func TestMoveUnknownMark(t *testing.T) {
	var l1 List[int]
	e1 := l1.PushBack(1)

	var l2 List[int]
	e2 := l2.PushBack(2)

	l1.MoveAfter(e1, e2)
	checkList(t, &l1, []int{1})
	checkList(t, &l2, []int{2})

	l1.MoveBefore(e1, e2)
	checkList(t, &l1, []int{1})
	checkList(t, &l2, []int{2})
}
