package csv

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestReader(t *testing.T) {
	type result struct {
		I   string `csv:"A"`
		II  int    `csv:"B"`
		III []int  `csv:"C"`
	}
	csv := `A,B,C
test
a,1,"[1,2]"
b,2,"[3,4]"
`
	r, err := NewReader(strings.NewReader(csv), true)
	if err != nil {
		t.Fatal(err)
	}
	if expect := []string{"A", "B", "C"}; !slices.Equal(expect, r.fields) {
		t.Errorf("expected %v; got %v", expect, r.fields)
	}

	var res1 []result
	for r.Next() {
		var res result
		if err := r.Scan(&res.I, &res.II, &res.III); err == nil {
			res1 = append(res1, res)
		}
	}
	if expect := []result{{"a", 1, []int{1, 2}}, {"b", 2, []int{3, 4}}}; !reflect.DeepEqual(expect, res1) {
		t.Errorf("expected %v; got %v", expect, res1)
	}

	r, err = NewReader(strings.NewReader(csv), true)
	if err != nil {
		t.Fatal(err)
	}
	var res2 []map[string]string
	for r.Next() {
		var res map[string]string
		if err := r.Decode(&res); err == nil {
			res2 = append(res2, res)
		}
	}
	if expect := []map[string]string{
		{"A": "a", "B": "1", "C": "[1,2]"},
		{"A": "b", "B": "2", "C": "[3,4]"},
	}; !reflect.DeepEqual(expect, res2) {
		t.Errorf("expected %v; got %v", expect, res2)
	}

	if err := DecodeAll(strings.NewReader(csv), &res2); err == nil {
		t.Error("expected non-nil error; got nil")
	}

	csv = `A,B,C
a,1,"[1,2]"
b,2,"[3,4]"
`
	var res3 []*result
	if err := DecodeAll(strings.NewReader(csv), &res3); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(res3)
	if res, expect := string(b), `[{"I":"a","II":1,"III":[1,2]},{"I":"b","II":2,"III":[3,4]}]`; res != expect {
		t.Errorf("expected %q; got %q", expect, res)
	}
}

func TestDecodeSlice(t *testing.T) {
	csv := `1
2
3
`
	var s1 []string
	if err := DecodeAll(strings.NewReader(csv), &s1); err != nil {
		t.Fatal(err)
	}
	if expect := []string{"1", "2", "3"}; !reflect.DeepEqual(expect, s1) {
		t.Errorf("expected %v; got %v", expect, s1)
	}
	var s2 []int
	if err := DecodeAll(strings.NewReader(csv), &s2); err != nil {
		t.Fatal(err)
	}
	if expect := []int{1, 2, 3}; !reflect.DeepEqual(expect, s2) {
		t.Errorf("expected %v; got %v", expect, s2)
	}
}
