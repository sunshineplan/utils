package loadbalance

import (
	"slices"
	"testing"
)

func TestRoundRobin(t *testing.T) {
	r1 := RoundRobin([]string{"a", "b", "c"}...)
	if r1.Len() != 3 {
		t.Fatalf("want 3, got %d", r1.Len())
	}
	var res []string
	for range 6 {
		res = append(res, r1.Next())
	}
	if expect := []string{"a", "b", "c", "a", "b", "c"}; !slices.Equal(res, expect) {
		t.Fatalf("want %v, got %v", expect, res)
	}
	res = nil

	r2 := WeightedRoundRobin([]Weighted[string]{{"a", 2}, {"b", 1}, {"c", 1}}...)
	if r2.Len() != 4 {
		t.Fatalf("want 4, got %d", r2.Len())
	}
	for range 12 {
		res = append(res, r2.Next())
	}
	if expect := []string{"a", "a", "b", "c", "a", "a", "b", "c", "a", "a", "b", "c"}; !slices.Equal(res, expect) {
		t.Fatalf("want %v, got %v", expect, res)
	}
	res = nil

	ring := r1.Link(r2)
	if ring.Len() != 7 {
		t.Fatalf("want 7, got %d", ring.Len())
	}
	for range 7 {
		res = append(res, ring.Next())
	}
	if expect := []string{"a", "b", "c", "a", "a", "b", "c"}; !slices.Equal(res, expect) {
		t.Fatalf("want %v, got %v", expect, res)
	}
}
