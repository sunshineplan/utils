// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package container

import (
	"container/ring"
	"fmt"
	"testing"
)

// For debugging - keep around.
func dump(r *ring.Ring) {
	if r == nil {
		fmt.Println("empty")
		return
	}
	i, n := 0, r.Len()
	for p := r; i < n; p = p.Next() {
		fmt.Printf("%4d: %p = {<- %p | %p ->}\n", i, p, p.Prev(), p.Next())
		i++
	}
	fmt.Println()
}

func verify[T int](t *testing.T, r *Ring[T], N int, sum int) {
	// Len
	n := r.Len()
	if n != N {
		t.Errorf("r.Len() == %d; expected %d", n, N)
	}

	// iteration
	n = 0
	s := 0
	r.Do(func(p *T) {
		n++
		if p != nil {
			s += int(*p)
		}
	})
	if n != N {
		t.Errorf("number of forward iterations == %d; expected %d", n, N)
	}
	if sum >= 0 && s != sum {
		t.Errorf("forward ring sum = %d; expected %d", s, sum)
	}

	if r == nil || r.r == nil {
		return
	}

	// connections
	if r.Next().r != nil {
		var p *ring.Ring // previous element
		for q := r.r; p == nil || q != r.r; q = q.Next() {
			if p != nil && p != q.Prev() {
				t.Errorf("prev = %p, expected q.prev = %p\n", p, q.Prev())
			}
			p = q
		}
		if p != r.Prev().r {
			t.Errorf("prev = %p, expected r.prev = %p\n", p, r.Prev())
		}
	}

	// Move
	if r.Move(0).r != r.r {
		t.Errorf("r.Move(0) != r")
	}
	if r.Move(N).r != r.r {
		t.Errorf("r.Move(%d) != r", N)
	}
	if r.Move(-N).r != r.r {
		t.Errorf("r.Move(%d) != r", -N)
	}
	for i := 0; i < 10; i++ {
		ni := N + i
		mi := ni % N
		if r.Move(ni).r != r.Move(mi).r {
			t.Errorf("r.Move(%d) != r.Move(%d)", ni, mi)
		}
		if r.Move(-ni).r != r.Move(-mi).r {
			t.Errorf("r.Move(%d) != r.Move(%d)", -ni, -mi)
		}
	}
}

func TestCornerCases(t *testing.T) {
	var (
		r0 = &Ring[int]{newMutex(), nil}
		r1 = Ring[int]{newMutex(), new(ring.Ring)}
	)
	// Basics
	verify(t, r0, 0, 0)
	verify(t, &r1, 1, 0)
	// Insert
	r1.Link(r0)
	verify(t, r0, 0, 0)
	verify(t, &r1, 1, 0)
	// Insert
	r1.Link(r0)
	verify(t, r0, 0, 0)
	verify(t, &r1, 1, 0)
	// Unlink
	r1.Unlink(0)
	verify(t, &r1, 1, 0)
}

func makeN(n int) *Ring[int] {
	r := NewRing[int](n)
	for i := 1; i <= n; i++ {
		r.Set(i)
		r = r.Next()
	}
	return r
}

func sumN(n int) int { return (n*n + n) / 2 }

func TestNew(t *testing.T) {
	for i := 0; i < 10; i++ {
		r := NewRing[int](i)
		verify(t, r, i, -1)
	}
	for i := 0; i < 10; i++ {
		r := makeN(i)
		verify(t, r, i, sumN(i))
	}
}

func TestLink1(t *testing.T) {
	r1a := makeN(1)
	var r1b = Ring[int]{newMutex(), &ring.Ring{}}
	r2a := r1a.Link(&r1b)
	verify(t, r2a, 2, 1)
	if r2a.r != r1a.r {
		t.Errorf("a) 2-element link failed")
	}

	r2b := r2a.Link(r2a.Next())
	verify(t, r2b, 2, 1)
	if r2b.r != r2a.Next().r {
		t.Errorf("b) 2-element link failed")
	}

	r1c := r2b.Link(r2b)
	verify(t, r1c, 1, 1)
	verify(t, r2b, 1, 0)
}

func TestLink2(t *testing.T) {
	var r0 = &Ring[int]{newMutex(), nil}
	r1a := NewRing[int](1)
	r1a.Set(42)
	r1b := NewRing[int](1)
	r1b.Set(77)
	r10 := makeN(10)

	r1a.Link(r0)
	verify(t, r1a, 1, 42)

	r1a.Link(r1b)
	verify(t, r1a, 2, 42+77)

	r10.Link(r0)
	verify(t, r10, 10, sumN(10))

	r10.Link(r1a)
	verify(t, r10, 12, sumN(10)+42+77)
}

func TestLink3(t *testing.T) {
	var r = Ring[int]{newMutex(), new(ring.Ring)}
	n := 1
	for i := 1; i < 10; i++ {
		n += i
		verify(t, r.Link(NewRing[int](i)), n, -1)
	}
}

func TestUnlink(t *testing.T) {
	r10 := makeN(10)
	s10 := r10.Move(6)

	sum10 := sumN(10)

	verify(t, r10, 10, sum10)
	verify(t, s10, 10, sum10)

	r0 := r10.Unlink(0)
	verify(t, r0, 0, 0)

	r1 := r10.Unlink(1)
	verify(t, r1, 1, 2)
	verify(t, r10, 9, sum10-2)

	r9 := r10.Unlink(9)
	verify(t, r9, 9, sum10-2)
	verify(t, r10, 9, sum10-2)
}

func TestLinkUnlink(t *testing.T) {
	for i := 1; i < 4; i++ {
		ri := NewRing[int](i)
		for j := 0; j < i; j++ {
			rj := ri.Unlink(j)
			verify(t, rj, j, -1)
			verify(t, ri, i-j, -1)
			ri.Link(rj)
			verify(t, ri, i, -1)
		}
	}
}

// Test that calling Move() on an empty Ring initializes it.
func TestMoveEmptyRing(t *testing.T) {
	var r = Ring[int]{newMutex(), &ring.Ring{}}

	r.Move(1)
	verify(t, &r, 1, 0)
}
