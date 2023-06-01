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
	c, buf := new(Counter), new(bytes.Buffer)
	buf.Write(data1)
	buf.Write(data2)
	r := c.AddReader(buf)
	io.ReadAll(r)
	if count := c.Load(); count != dataLen {
		t.Fatalf("expected %d; got %d", dataLen, count)
	}
}

func TestWriter(t *testing.T) {
	c, buf := new(Counter), new(bytes.Buffer)
	w := c.AddWriter(buf)
	w.Write(data1)
	w.Write(data2)
	if count := c.Load(); count != dataLen {
		t.Fatalf("expected %d; got %d", dataLen, count)
	}
}
