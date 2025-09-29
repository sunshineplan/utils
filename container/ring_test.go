// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package container

import (
	"fmt"
	"sync"
	"testing"
)

// For debugging - keep around.
func dump[T any](r *Ring[T]) {
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

func verify(t *testing.T, r *Ring[int], N int, sum int) {
	// Len
	n := r.Len()
	if n != N {
		t.Errorf("r.Len() == %d; expected %d", n, N)
	}

	// iteration
	n = 0
	s := 0
	r.Do(func(p int) {
		n++
		s += p
	})
	if n != N {
		t.Errorf("number of forward iterations == %d; expected %d", n, N)
	}
	if sum >= 0 && s != sum {
		t.Errorf("forward ring sum = %d; expected %d", s, sum)
	}

	if r == nil {
		return
	}

	// connections
	if r.Next() != nil {
		var p *Ring[int] // previous element
		for q := r; p == nil || q != r; q = q.next {
			if p != nil && p != q.Prev() {
				t.Errorf("prev = %p, expected q.prev = %p\n", p, q.Prev())
			}
			p = q
		}
		if p != r.Prev() {
			t.Errorf("prev = %p, expected r.prev = %p\n", p, r.Prev())
		}
	}

	// Move
	if r.Move(0) != r {
		t.Errorf("r.Move(0) != r")
	}
	if r.Move(N) != r {
		t.Errorf("r.Move(%d) != r", N)
	}
	if r.Move(-N) != r {
		t.Errorf("r.Move(%d) != r", -N)
	}
	for i := 0; i < 10; i++ {
		ni := N + i
		mi := ni % N
		if r.Move(ni) != r.Move(mi) {
			t.Errorf("r.Move(%d) != r.Move(%d)", ni, mi)
		}
		if r.Move(-ni) != r.Move(-mi) {
			t.Errorf("r.Move(%d) != r.Move(%d)", -ni, -mi)
		}
	}
}

func TestCornerCases(t *testing.T) {
	var (
		r0 *Ring[int]
		r1 Ring[int]
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
	var r1b Ring[int]
	r2a := r1a.Link(&r1b)
	verify(t, r2a, 2, 1)
	if r2a != r1a {
		t.Errorf("a) 2-element link failed")
	}

	r2b := r2a.Link(r2a.Next())
	verify(t, r2b, 2, 1)
	if r2b != r2a.Next() {
		t.Errorf("b) 2-element link failed")
	}

	r1c := r2b.Link(r2b)
	verify(t, r1c, 1, 1)
	verify(t, r2b, 1, 0)
}

func TestLink2(t *testing.T) {
	var r0 *Ring[int]
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
	var r Ring[int]
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
	var r Ring[int]

	r.Move(1)
	verify(t, &r, 1, 0)
}

// TestLinkSharedRing tests the Link function to ensure the shared ringMu is correctly set.
func TestLinkSharedRing(t *testing.T) {
	// Helper function to check if all elements in a ring share the same ringMu.
	checkRingMu := func(t *testing.T, r *Ring[int], expectedMu *sync.RWMutex, name string) {
		if r == nil {
			t.Errorf("%s: ring is nil", name)
			return
		}
		seen := make(map[*Ring[int]]bool)
		current := r
		for {
			if current.ringMu != expectedMu {
				t.Errorf("%s: element %p has ringMu %p, expected %p", name, current, current.ringMu, expectedMu)
			}
			seen[current] = true
			current = current.Next()
			if current == r {
				break
			}
			if seen[current] {
				t.Errorf("%s: cycle detected before reaching start", name)
				break
			}
		}
	}

	// Test 1: Link two distinct rings.
	t.Run("LinkDistinctRings", func(t *testing.T) {
		r1 := NewRing[int](3)
		r2 := NewRing[int](2)
		originalR1Mu := r1.ringMu
		originalR2Mu := r2.ringMu

		// Link r1 and r2.
		r1.Link(r2)

		if originalR1Mu == originalR2Mu {
			t.Fatalf("initial rings have same ringMu %p", originalR1Mu)
		}

		// Check that all elements in the combined ring share r1's ringMu.
		checkRingMu(t, r1, originalR1Mu, "combined ring")

		// Verify ring length.
		if r1.Len() != 5 {
			t.Errorf("combined ring length is %d, expected 5", r1.Len())
		}
	})

	// Test 2: Link within the same ring.
	t.Run("LinkSameRing", func(t *testing.T) {
		r := NewRing[int](5)
		originalMu := r.ringMu

		// Move to the third element.
		r3 := r.Move(2)

		// Link r to r3, splitting the ring.
		result := r.Link(r3)

		// Check that the original ring (r) has 2 elements and retains original ringMu.
		checkRingMu(t, r, originalMu, "original ring")

		// Check that the result ring has 3 elements and a new ringMu.
		if result.ringMu == originalMu {
			t.Errorf("result ring has same ringMu %p as original %p", result.ringMu, originalMu)
		}

		checkRingMu(t, result, result.ringMu, "result ring")
	})

	// Test 3: Link a ring to itself.
	t.Run("LinkSelf", func(t *testing.T) {
		r := NewRing[int](5)
		originalMu := r.ringMu

		result := r.Link(r)

		// Check that the ring remains unchanged and retains original ringMu.
		checkRingMu(t, r, originalMu, "self-linked ring")
		if r.Len() != 1 {
			t.Errorf("self-linked ring length is %d, expected 1", r.Len())
		}

		// Check that the result ring has 3 elements and a new ringMu.
		if result.ringMu == originalMu {
			t.Errorf("result ring has same ringMu %p as original %p", result.ringMu, originalMu)
		}

		checkRingMu(t, result, result.ringMu, "result ring")
	})

	// Test 4: Link with nil.
	t.Run("LinkNil", func(t *testing.T) {
		r := NewRing[int](3)
		originalMu := r.ringMu
		originalNext := r.Next()

		result := r.Link(nil)

		// Check that the ring is unchanged and retains original ringMu.
		checkRingMu(t, r, originalMu, "ring after linking nil")
		if r.Len() != 3 {
			t.Errorf("ring length is %d, expected 3", r.Len())
		}
		if result != originalNext {
			t.Errorf("Link result is %p, expected %p", result, originalNext)
		}
	})
}
