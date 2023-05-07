package counter

import (
	"net"
	"sync/atomic"
)

var (
	_ net.Listener = &Listener{}
	_ net.Conn     = &conn{}
)

type Listener struct {
	l       net.Listener
	read    atomic.Int64
	written atomic.Int64
}

func NewListener(l net.Listener) *Listener {
	return &Listener{l: l}
}

func (l *Listener) Accept() (net.Conn, error) {
	c, err := l.l.Accept()
	if err != nil {
		return nil, err
	}
	return &conn{Conn: c, l: l}, nil
}

func (l *Listener) Close() error {
	return l.l.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.l.Addr()
}

func (l *Listener) ReadCount() int64 {
	return l.read.Load()
}

func (l *Listener) WriteCount() int64 {
	return l.written.Load()
}

type conn struct {
	net.Conn
	l *Listener
}

func (conn *conn) Write(b []byte) (n int, err error) {
	n, err = conn.Conn.Write(b)
	if err != nil {
		return
	}
	conn.l.written.Add(int64(n))
	return
}

func (conn *conn) Read(b []byte) (n int, err error) {
	n, err = conn.Conn.Read(b)
	if err != nil {
		return
	}
	conn.l.read.Add(int64(n))
	return
}
