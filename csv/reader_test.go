package csv

import (
	"reflect"
	"strings"
	"testing"
)

type result struct {
	A string
	B int
	C []int
}

func TestReader(t *testing.T) {
	csv := `A,B,C
test
a,1,"[1,2]"
b,2,"[3,4]"
`
	rs := FromReader(strings.NewReader(csv))

	fields, err := rs.Read()
	if err != nil {
		t.Fatal(err)
	}
	if expect := []string{"A", "B", "C"}; !reflect.DeepEqual(expect, fields) {
		t.Errorf("expected %v; got %v", expect, fields)
	}

	var results []result
	for rs.Next() {
		var result result
		if err := rs.Scan(&result.A, &result.B, &result.C); err == nil {
			results = append(results, result)
		}
	}
	if expect := []result{{"a", 1, []int{1, 2}}, {"b", 2, []int{3, 4}}}; !reflect.DeepEqual(expect, results) {
		t.Errorf("expected %v; got %v", expect, results)
	}
}
