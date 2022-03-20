package csv

import (
	"bytes"
	"testing"
)

type testcase[T any] struct {
	name       string
	fieldnames []string
	slice      []T
}

func testExport[T any](t *testing.T, tc testcase[T], result string) {
	var b bytes.Buffer
	if err := Export(tc.fieldnames, tc.slice, &b); err != nil {
		t.Error(tc.name, err)
	}
	if r := b.String(); r != result {
		t.Errorf("%s expected %q; got %q", tc.name, result, r)
	}
}

func TestExport(t *testing.T) {
	type test struct{ A, B any }
	result := `A,B
a,b
aa,
`
	testExport(t, testcase[map[string]any]{
		name:       "map slice",
		fieldnames: []string{"A", "B"},
		slice: []map[string]any{
			{"A": "a", "B": "b"},
			{"A": "aa", "B": nil},
		},
	}, result)
	testExport(t, testcase[*map[string]any]{
		name:       "pointer map slice",
		fieldnames: []string{"A", "B"},
		slice: []*map[string]any{
			{"A": "a", "B": "b"},
			{"A": "aa", "B": nil},
		},
	}, result)
	testExport(t, testcase[test]{
		name:       "struct slice",
		fieldnames: []string{"A", "B"},
		slice:      []test{{A: "a", B: "b"}, {A: "aa", B: nil}},
	}, result)
	testExport(t, testcase[test]{
		name:       "struct slice without fieldnames",
		fieldnames: nil,
		slice:      []test{{A: "a", B: "b"}, {A: "aa", B: nil}},
	}, result)
	testExport(t, testcase[*test]{
		name:       "pointer struct slice without fieldnames",
		fieldnames: nil,
		slice:      []*test{{A: "a", B: "b"}, nil, {A: "aa", B: nil}},
	}, result)
	testExport(t, testcase[D]{
		name:       "D slice",
		fieldnames: []string{"A", "B"},
		slice:      []D{{{"A", "a"}, {"B", "b"}}, {{"A", "aa"}, {"B", nil}}},
	}, result)
	testExport(t, testcase[D]{
		name:       "D slice without fieldnames",
		fieldnames: nil,
		slice:      []D{{{"A", "a"}, {"B", "b"}}, {{"A", "aa"}, {"B", nil}}},
	}, result)
	testExport(t, testcase[any]{
		name:       "interface slice",
		fieldnames: []string{"A", "B"},
		slice: []any{
			test{A: "a", B: "b"},
			map[string]any{"A": "aa", "B": nil},
		},
	}, result)
}

func TestExportStruct(t *testing.T) {
	type test struct {
		A string `csv:"a"`
		B []int
	}
	result := `a,B
a,"[1,2]"
`

	var b bytes.Buffer
	if err := Export(nil, []test{{A: "a", B: []int{1, 2}}}, &b); err != nil {
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
	if err := ExportUTF8([]string{"A", "B"}, []any{map[string]string{"A": "a", "B": "b"}}, &b); err != nil {
		t.Error(err)
	}
	c := b.Bytes()
	if !bytes.Equal(utf8bom, c[:3]) {
		t.Errorf("expected %q; got %q", utf8bom, c[:3])
	}
	if r := string(c[3:]); r != result {
		t.Errorf("expected %q; got %q", result, r)
	}
}
