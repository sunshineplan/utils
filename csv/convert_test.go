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
	if s != "string" {
		t.Errorf("expected %q; got %q", "string", s)
	}

	var n int
	if err := convertAssign(&n, "123"); err != nil {
		t.Fatal(err)
	}
	if n != 123 {
		t.Errorf("expected %d; got %d", 123, n)
	}

	var a []int
	if err := convertAssign(&a, "[1,2]"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]int{1, 2}, a) {
		t.Errorf("expected %v; got %v", []int{1, 2}, a)
	}
}
