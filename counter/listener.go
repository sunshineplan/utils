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
	net.Listener
	readBytes  Counter
	writeBytes Counter
}

func NewListener(listener net.Listener) *Listener {
	return &Listener{Listener: listener}
}

func (l *Listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &conn{
		Conn: c,
		r:    CountReader(c, &l.readBytes),
		w:    CountWriter(c, &l.writeBytes),
	}, nil
}

func (l *Listener) ReadBytes() int64 {
	return l.readBytes.Get()
}

func (l *Listener) WriteBytes() int64 {
	return l.writeBytes.Get()
}

type conn struct {
	net.Conn
	r io.Reader
	w io.Writer
}

func (c *conn) Read(b []byte) (int, error)  { return c.r.Read(b) }
func (c *conn) Write(b []byte) (int, error) { return c.w.Write(b) }
