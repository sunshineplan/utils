package counter

import (
	"bytes"
	"testing"
)

func TestWriterCounter(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.Write(data1)
	w.Write(data2)
	if count := w.Count(); count != dataLen {
		t.Fatalf("expected %d; got %d", dataLen, count)
	}
}
