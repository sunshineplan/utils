package slice

import (
	"reflect"
	"testing"
)

func testDeduplicate[E comparable](t *testing.T, s1, s2 []E) {
	res := Deduplicate(s1)
	if !reflect.DeepEqual(res, s2) {
		t.Errorf("expected %#v; got %#v", s2, res)
	}
}

func TestDeduplicate(t *testing.T) {
	type test struct {
		a, b string
	}

	testDeduplicate(t, []test{{"a", "b"}, {"a", "b"}, {"b", "c"}}, []test{{"a", "b"}, {"b", "c"}})
	testDeduplicate(t, []int{1, 2, 2, 3}, []int{1, 2, 3})
	testDeduplicate(t, []string{"a", "b", "b", "c"}, []string{"a", "b", "c"})
	testDeduplicate(t, []test{}, []test{})
	testDeduplicate(t, []test(nil), []test(nil))

	res := Deduplicate([]test{{"a", "b"}, {"a", "b"}, {"b", "c"}})
	if l := len(res); l != 2 {
		t.Errorf("expected %v; got %v", 2, l)
	}
}
