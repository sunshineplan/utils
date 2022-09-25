package loadbalance

import (
	"reflect"
	"testing"
)

func TestRoundRobin(t *testing.T) {
	if _, err := RoundRobin[struct{}](); err == nil {
		t.Error("want error, got nil")
	}

	a, b, c := new(string), new(string), new(string)
	*a, *b, *c = "a", "b", "c"

	loadbalancer, err := RoundRobin([]*string{a, b, c}...)
	if err != nil {
		t.Error(err)
	} else {
		var res []*string
		for i := 0; i < 6; i++ {
			res = append(res, loadbalancer.Next())
		}
		if expect := []*string{a, b, c, a, b, c}; !reflect.DeepEqual(res, expect) {
			t.Errorf("want %v, got %v", expect, res)
		}
	}

	loadbalancer, err = WeightedRoundRobin([]Weighted[string]{{a, 2}, {b, 1}, {c, 1}}...)
	if err != nil {
		t.Error(err)
	} else {
		var res []*string
		for i := 0; i < 8; i++ {
			res = append(res, loadbalancer.Next())
		}
		if expect := []*string{a, a, b, c, a, a, b, c}; !reflect.DeepEqual(res, expect) {
			t.Errorf("want %v, got %v", expect, res)
		}
	}
}
