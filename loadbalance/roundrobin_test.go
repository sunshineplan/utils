package loadbalance

import (
	"reflect"
	"testing"
)

func TestRoundRobin(t *testing.T) {
	r1 := RoundRobin([]string{"a", "b", "c"}...)
	var res []string
	for range 6 {
		res = append(res, r1.Next())
	}
	if expect := []string{"a", "b", "c", "a", "b", "c"}; !reflect.DeepEqual(res, expect) {
		t.Errorf("want %v, got %v", expect, res)
	}
	res = nil

	r2 := WeightedRoundRobin([]Weighted[string]{{"a", 2}, {"b", 1}, {"c", 1}}...)
	for range 12 {
		res = append(res, r2.Next())
	}
	if expect := []string{"a", "a", "b", "c", "a", "a", "b", "c", "a", "a", "b", "c"}; !reflect.DeepEqual(res, expect) {
		t.Errorf("want %v, got %v", expect, res)
	}
	res = nil

	r1.Link(r2)
	for range 7 {
		res = append(res, r1.Next())
	}
	if expect := []string{"a", "a", "a", "b", "c", "b", "c"}; !reflect.DeepEqual(res, expect) {
		t.Errorf("want %v, got %v", expect, res)
	}
}
