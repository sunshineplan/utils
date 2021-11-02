package counter

import (
	"bytes"
	"testing"
)

var data = []byte("Hello, World!")
var dataLen = uint64(len(data))

func TestWriterCounter(t *testing.T) {
	buf := bytes.Buffer{}
	w := NewWriter(&buf)
	w.Write(data)
	w.Write(data)
	if w.Count() != dataLen*2 {
		t.Fatalf("count mismatch len of test data: %d != %d", w.Count(), len(data)*2)
	}
}
