package counter

import (
	"io"
	"net"
)

var (
	_ net.Listener = &Listener{}
	_ net.Conn     = &conn{}
)

type Listener struct {
	listener net.Listener
	read     Counter
	written  Counter
}

func NewListener(listener net.Listener) *Listener {
	return &Listener{listener: listener}
}

func (l *Listener) Accept() (net.Conn, error) {
	c, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &conn{
		Conn: c,
		r:    l.read.AddReader(c),
		w:    l.written.AddWriter(c),
	}, nil
}

func (l *Listener) Close() error {
	return l.listener.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *Listener) ReadCount() int64 {
	return l.read.Load()
}

func (l *Listener) WriteCount() int64 {
	return l.written.Load()
}

type conn struct {
	net.Conn
	r io.Reader
	w io.Writer
}

func (c *conn) Read(b []byte) (int, error)  { return c.r.Read(b) }
func (c *conn) Write(b []byte) (int, error) { return c.w.Write(b) }
