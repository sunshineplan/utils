package counter

import (
	"bytes"
	"io"
	"testing"
)

var data1 = []byte("Hello, World!")
var data2 = []byte("Hello, Golang!")
var dataLen = int64(len(data1) + len(data2))

func TestReader(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(data1)
	buf.Write(data2)
	r := NewReader(&buf)
	io.ReadAll(r)
	if count := r.Count(); count != dataLen {
		t.Fatalf("expected %d; got %d", dataLen, count)
	}
}
