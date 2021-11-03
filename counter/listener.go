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
	net.Listener
	read    uint64
	written uint64
}

func NewListener(l net.Listener) *Listener {
	return &Listener{Listener: l}
}

func (l *Listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &conn{Conn: c, l: l}, nil
}

func (l *Listener) ReadCount() uint64 {
	return atomic.LoadUint64(&l.read)
}

func (l *Listener) WriteCount() uint64 {
	return atomic.LoadUint64(&l.written)
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
	atomic.AddUint64(&conn.l.written, uint64(n))
	return
}

func (conn *conn) Read(b []byte) (n int, err error) {
	n, err = conn.Conn.Read(b)
	if err != nil {
		return
	}
	atomic.AddUint64(&conn.l.read, uint64(n))
	return
}
