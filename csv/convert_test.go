package csv

import (
	"reflect"
	"testing"
)

func TestConvert(t *testing.T) {
	var s string
	if err := convertAssign(&s, "string"); err != nil {
		t.Fatal(err)
	}
	if expect := "string"; s != expect {
		t.Errorf("expected %q; got %q", expect, s)
	}

	var n int
	if err := convertAssign(&n, "123"); err != nil {
		t.Fatal(err)
	}
	if expect := 123; n != expect {
		t.Errorf("expected %d; got %d", expect, n)
	}

	var a []int
	if err := convertAssign(&a, "[1,2]"); err != nil {
		t.Fatal(err)
	}
	if expect := []int{1, 2}; !reflect.DeepEqual(expect, a) {
		t.Errorf("expected %v; got %v", expect, a)
	}
}
