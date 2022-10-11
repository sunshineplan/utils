package csv

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

type test string

func (v *test) UnmarshalText(b []byte) (err error) {
	if len(b) == 0 {
		return nil
	}
	var s []string
	if err = json.Unmarshal(b, &s); err == nil {
		*v = test(strings.Join(s, ","))
	}
	return
}

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

	var ts test
	if err := convertAssign(&ts, ""); err != nil {
		t.Fatal(err)
	}
	if expect := test(""); expect != ts {
		t.Errorf("expected %v; got %v", expect, ts)
	}
	if err := convertAssign(&ts, `["1","2"]`); err != nil {
		t.Fatal(err)
	}
	if expect := test("1,2"); expect != ts {
		t.Errorf("expected %v; got %v", expect, ts)
	}
}
