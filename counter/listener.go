package counter

import "net"

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
	return &conn{c, l}, nil
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
	listener *Listener
}

func (conn *conn) Write(b []byte) (n int, err error) {
	return conn.listener.written.AddWriter(conn.Conn).Write(b)
}

func (conn *conn) Read(b []byte) (n int, err error) {
	return conn.listener.read.AddReader(conn.Conn).Read(b)
}
