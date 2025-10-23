package counter

import (
	"io"
	"net"
	"os"
	"testing"
)

func TestListener(t *testing.T) {
	listener, err := net.Listen("unix", "tmp.sock")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("tmp.sock")

	l := NewListener(listener)

	conn, err := net.Dial(l.Addr().Network(), l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn.Write([]byte(data1))
	conn.Close()

	conn, err = l.Accept()
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.ReadAll(conn)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()

	conn, err = net.Dial(l.Addr().Network(), l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn.Write([]byte(data2))
	conn.Close()

	conn, err = l.Accept()
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.ReadAll(conn)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()

	if n := l.ReadBytes(); n != dataLen {
		t.Fatalf("expected %d; got %d", dataLen, n)
	}
}
