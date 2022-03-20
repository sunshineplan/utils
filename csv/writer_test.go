package csv

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestWriteFields(t *testing.T) {
	w := NewWriter(io.Discard, false)
	if err := w.Write(map[string]string{"test": "test"}); err == nil {
		t.Error("gave nil error; want error")
	}
	if err := w.WriteFields(map[string]string{"test": "test"}); err == nil {
		t.Error("gave nil error; want error")
	}
	if err := w.WriteFields(struct {
		A string
		B string `csv:"b"`
	}{}); err != nil {
		t.Error(err)
	} else {
		if expect := []field{{"A", ""}, {"B", "b"}}; !reflect.DeepEqual(expect, w.fields) {
			t.Errorf("expected %v; got %v", expect, w.fields)
		}
	}
}

func TestWriter(t *testing.T) {
	result := `A|B
a|b
aa|
`

	var buf bytes.Buffer
	w := NewWriter(&buf, false)
	w.Comma = '|'
	if err := w.WriteFields(struct{ A, B any }{}); err != nil {
		t.Fatal(err)
	}
	if err := w.Write(struct{ A, B any }{}); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteAll([]struct{ A, B any }{{A: "a", B: "b"}, {A: "aa", B: nil}}); err != nil {
		t.Fatal(err)
	}
	if r := buf.String(); r != result {
		t.Errorf("expected %q; got %q", result, r)
	}
}
