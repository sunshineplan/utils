package csv

import (
	"io"
	"reflect"
	"testing"
)

func TestWriteFields(t *testing.T) {
	w := NewWriter(io.Discard, false)
	if err := w.WriteFields(map[string]string{"test": "test"}); err == nil {
		t.Error("gave nil error; want error")
	}
	if err := w.WriteFields(struct{ A, B string }{}); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual([]string{"A", "B"}, w.fields) {
			t.Errorf("expected %q; got %q", []string{"A", "B"}, w.fields)
		}
	}
}
