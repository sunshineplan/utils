package utils

import (
	"reflect"
	"testing"
)

func TestDeduplicate(t *testing.T) {
	type test struct {
		a, b string
	}
	tc := []struct {
		slice, unique interface{}
	}{
		{[]test{{"a", "b"}, {"a", "b"}, {"b", "c"}}, []test{{"a", "b"}, {"b", "c"}}},
		{[]int{1, 2, 2, 3}, []int{1, 2, 3}},
		{[]string{"a", "b", "b", "c"}, []string{"a", "b", "c"}},
		{[]test{}, []test{}},
	}
	for _, i := range tc {
		unique := Deduplicate(i.slice)
		if !reflect.DeepEqual(unique, i.unique) {
			t.Errorf("expected %v; got %v", i.unique, unique)
		}
	}

	unique := Deduplicate([]test{{"a", "b"}, {"a", "b"}, {"b", "c"}})
	if l := len(unique.([]test)); l != 2 {
		t.Errorf("expected %v; got %v", 2, l)
	}
}
