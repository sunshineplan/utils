package container

import (
	"sync"
	"testing"
)

func TestMap_BasicOperations(t *testing.T) {
	m := NewMap[string, int]()

	// Store & Load
	m.Store("a", 1)
	if v, ok := m.Load("a"); !ok || v != 1 {
		t.Fatalf("Load failed, got (%v, %v), want (1, true)", v, ok)
	}

	// LoadOrStore (existing key)
	actual, loaded := m.LoadOrStore("a", 2)
	if !loaded || actual != 1 {
		t.Fatalf("LoadOrStore existing key failed, got (%v, %v), want (1, true)", actual, loaded)
	}

	// LoadOrStore (new key)
	actual, loaded = m.LoadOrStore("b", 3)
	if loaded || actual != 3 {
		t.Fatalf("LoadOrStore new key failed, got (%v, %v), want (3, false)", actual, loaded)
	}

	// LoadAndDelete
	v, ok := m.LoadAndDelete("b")
	if !ok || v != 3 {
		t.Fatalf("LoadAndDelete failed, got (%v, %v), want (3, true)", v, ok)
	}
	if _, ok := m.Load("b"); ok {
		t.Fatalf("Key 'b' should be deleted")
	}

	// Swap
	prev, loaded := m.Swap("a", 5)
	if !loaded || prev != 1 {
		t.Fatalf("Swap failed, got (%v, %v), want (1, true)", prev, loaded)
	}
	v, ok = m.Load("a")
	if !ok || v != 5 {
		t.Fatalf("Swap did not update value, got (%v, %v), want (5, true)", v, ok)
	}

	// CompareAndSwap
	if !m.CompareAndSwap("a", 5, 10) {
		t.Fatalf("CompareAndSwap should succeed")
	}
	v, _ = m.Load("a")
	if v != 10 {
		t.Fatalf("CompareAndSwap failed, value = %v, want 10", v)
	}

	if m.CompareAndSwap("a", 5, 20) {
		t.Fatalf("CompareAndSwap should fail when old != current")
	}

	// CompareAndDelete
	m.Store("x", 42)
	if !m.CompareAndDelete("x", 42) {
		t.Fatalf("CompareAndDelete should succeed")
	}
	if _, ok := m.Load("x"); ok {
		t.Fatalf("CompareAndDelete did not delete key")
	}

	m.Store("y", 99)
	if m.CompareAndDelete("y", 100) {
		t.Fatalf("CompareAndDelete should fail when old != current")
	}
}

func TestMap_Range(t *testing.T) {
	m := NewMap[string, int]()
	m.Store("a", 1)
	m.Store("b", 2)
	m.Store("c", 3)

	collected := make(map[string]int)
	m.Range(func(k string, v int) bool {
		collected[k] = v
		return true
	})

	if len(collected) != 3 {
		t.Fatalf("Range failed, got %d elements, want 3", len(collected))
	}
}

func TestMap_Clear(t *testing.T) {
	m := NewMap[string, int]()
	m.Store("a", 1)
	m.Store("b", 2)
	m.Clear()

	count := 0
	m.Range(func(k string, v int) bool {
		count++
		return true
	})

	if count != 0 {
		t.Fatalf("Clear failed, map still has %d elements", count)
	}
}

func TestMap_ConcurrentAccess(t *testing.T) {
	m := NewMap[int, int]()
	wg := sync.WaitGroup{}

	// Concurrent store
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.Store(i, i*i)
		}(i)
	}

	wg.Wait()

	// Concurrent load
	wg = sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v, ok := m.Load(i)
			if !ok {
				t.Errorf("Missing key %d", i)
			} else if v != i*i {
				t.Errorf("Unexpected value for %d: got %d, want %d", i, v, i*i)
			}
		}(i)
	}
	wg.Wait()
}
