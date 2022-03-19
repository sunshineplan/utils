package csv

import (
	"reflect"
	"strings"
	"testing"
)

func TestReader(t *testing.T) {
	type result struct {
		A string
		B int
		C []int
	}
	csv := `A,B,C
test
a,1,"[1,2]"
b,2,"[3,4]"
`
	r := NewReader(strings.NewReader(csv), true)
	if expect := []string{"A", "B", "C"}; !reflect.DeepEqual(expect, r.fields) {
		t.Errorf("expected %v; got %v", expect, r.fields)
	}

	var results []result
	for r.Next() {
		var result result
		if err := r.Scan(&result.A, &result.B, &result.C); err == nil {
			results = append(results, result)
		}
	}
	if expect := []result{{"a", 1, []int{1, 2}}, {"b", 2, []int{3, 4}}}; !reflect.DeepEqual(expect, results) {
		t.Errorf("expected %v; got %v", expect, results)
	}

	r = NewReader(strings.NewReader(csv), true)
	results = nil
	for r.Next() {
		var result result
		if err := r.Decode(&result); err == nil {
			results = append(results, result)
		}
	}
	if expect := []result{{"a", 1, []int{1, 2}}, {"b", 2, []int{3, 4}}}; !reflect.DeepEqual(expect, results) {
		t.Errorf("expected %v; got %v", expect, results)
	}

	csv = `A,B,C
a,1,"[1,2]"
b,2,"[3,4]"
`
	results = nil
	if err := DecodeAll(strings.NewReader(csv), &results); err != nil {
		t.Fatal(err)
	}
	if expect := []result{{"a", 1, []int{1, 2}}, {"b", 2, []int{3, 4}}}; !reflect.DeepEqual(expect, results) {
		t.Errorf("expected %v; got %v", expect, results)
	}
}
