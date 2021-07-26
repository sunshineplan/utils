package csv

import (
	"bytes"
	"reflect"
	"testing"
)

func TestExport(t *testing.T) {
	type test struct{ A, B interface{} }
	testcase := []struct {
		name       string
		fieldnames []string
		slice      interface{}
	}{
		{
			name:       "map slice",
			fieldnames: []string{"A", "B"},
			slice: []map[string]interface{}{
				{"A": "a", "B": "b"},
				{"A": "aa", "B": nil},
			},
		},
		{
			name:       "struct slice",
			fieldnames: []string{"A", "B"},
			slice:      []test{{A: "a", B: "b"}, {A: "aa", B: nil}},
		},
		{
			name:       "struct slice without fieldnames",
			fieldnames: nil,
			slice:      []test{{A: "a", B: "b"}, {A: "aa", B: nil}},
		},
		{
			name:       "interface slice",
			fieldnames: []string{"A", "B"},
			slice: []interface{}{
				test{A: "a", B: "b"},
				map[string]interface{}{"A": "aa", "B": nil},
			},
		},
	}
	result := `A,B
a,b
aa,
`

	for _, tc := range testcase {
		var b bytes.Buffer
		if err := Export(tc.fieldnames, tc.slice, &b); err != nil {
			t.Error(tc.name, err)
		}
		if r := b.String(); r != result {
			t.Errorf("%s expected %q; got %q", tc.name, result, r)
		}
	}
}

func TestExportStruct(t *testing.T) {
	type test struct {
		A string
		B []int
	}
	result := `A,B
a,"[1,2]"
`

	var b bytes.Buffer
	if err := Export([]string{"A", "B"}, []test{{A: "a", B: []int{1, 2}}}, &b); err != nil {
		t.Fatal(err)
	}
	if r := b.String(); r != result {
		t.Errorf("expected %q; got %q", result, r)
	}
}

func TestExportUTF8(t *testing.T) {
	result := `A,B
a,b
`
	var b bytes.Buffer
	if err := ExportUTF8([]string{"A", "B"}, []interface{}{map[string]string{"A": "a", "B": "b"}}, &b); err != nil {
		t.Error(err)
	}
	c := b.Bytes()
	if !reflect.DeepEqual(utf8bom, c[:3]) {
		t.Errorf("expected %q; got %q", utf8bom, c[:3])
	}
	if r := string(c[3:]); r != result {
		t.Errorf("expected %q; got %q", result, r)
	}
}
