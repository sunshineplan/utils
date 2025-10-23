package counter

import (
	"io"
	"net"
)

var (
	_ net.Listener = &Listener{}
	_ net.Conn     = &conn{}
)

// Listener wraps a net.Listener to count bytes read and written across all connections.
type Listener struct {
	net.Listener
	readBytes  Counter // Counter for bytes read across all connections
	writeBytes Counter // Counter for bytes written across all connections
}

// NewListener creates a Listener that counts bytes read and written across all connections.
func NewListener(listener net.Listener) *Listener {
	return &Listener{Listener: listener}
}

// Accept accepts a connection and wraps it with byte counting for reads and writes.
// It returns the wrapped connection or an error if the accept fails.
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

// ReadBytes returns the total number of bytes read across all connections.
func (l *Listener) ReadBytes() int64 {
	return l.readBytes.Get()
}

// WriteBytes returns the total number of bytes written across all connections.
func (l *Listener) WriteBytes() int64 {
	return l.writeBytes.Get()
}

// conn wraps a net.Conn to count bytes read and written.
type conn struct {
	net.Conn
	r io.Reader // Reader that counts bytes read
	w io.Writer // Writer that counts bytes written
}

// Read reads from the underlying Reader and counts the bytes read.
// It returns the number of bytes read and any error encountered.
func (c *conn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

// Write writes to the underlying Writer and counts the bytes written.
// It returns the number of bytes written and any error encountered.
func (c *conn) Write(b []byte) (int, error) {
	return c.w.Write(b)
}
