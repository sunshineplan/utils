package loadbalance

import (
	"reflect"
	"sort"
	"testing"
)

func TestRandom(t *testing.T) {
	if _, err := Random[struct{}](); err == nil {
		t.Error("want error, got nil")
	}

	a, b, c := "a", "b", "c"

	loadbalancer, err := Random([]string{a, b, c}...)
	if err != nil {
		t.Error(err)
	} else {
		var res []string
		for range 6 {
			res = append(res, loadbalancer.Next())
		}
		sort.Strings(res)
		if expect := []string{a, a, b, b, c, c}; !reflect.DeepEqual(res, expect) {
			t.Errorf("want %v, got %v", expect, res)
		}
	}

	loadbalancer, err = WeightedRandom([]*Weighted[string]{{a, 2}, {b, 1}, {c, 1}}...)
	if err != nil {
		t.Error(err)
	} else {
		var res []string
		for range 8 {
			res = append(res, loadbalancer.Next())
		}
		sort.Strings(res)
		if expect := []string{a, a, a, a, b, b, c, c}; !reflect.DeepEqual(res, expect) {
			t.Errorf("want %v, got %v", expect, res)
		}
	}
}
