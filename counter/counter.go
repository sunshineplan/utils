package counter

import (
	"io"
	"sync/atomic"
)

// Counter is a thread-safe utility for counting values, starting from zero.
type Counter struct {
	n atomic.Int64
}

// Add adds delta to the [Counter] and returns the new value.
func (c *Counter) Add(delta int64) int64 {
	return c.n.Add(delta)
}

// Get returns the current value of the [Counter].
func (c *Counter) Get() int64 {
	return c.n.Load()
}

// CounterReader wraps an io.Reader to count the number of bytes read.
type CounterReader struct {
	r io.Reader // Underlying reader
	c *Counter  // Counter for bytes read
}

// CountReader creates an io.Reader that counts bytes read from r, using the provided [Counter].
func CountReader(r io.Reader, c *Counter) io.Reader {
	return NewCounterReader(r, c)
}

// NewCounterReader creates a [CounterReader] that counts bytes read from r, using the provided [Counter].
// If c is nil, a new Counter is created.
func NewCounterReader(r io.Reader, c *Counter) *CounterReader {
	if c == nil {
		c = new(Counter)
	}
	return &CounterReader{r, c}
}

// Read reads from the underlying Reader and increments the counter by the number of bytes read.
// It returns the number of bytes read and any error encountered.
func (r *CounterReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	if n > 0 {
		r.c.Add(int64(n))
	}
	return
}

// Bytes returns the total number of bytes read.
func (r *CounterReader) Bytes() int64 {
	return r.c.Get()
}

// CounterWriter wraps an io.Writer to count the number of bytes written.
type CounterWriter struct {
	w io.Writer // Underlying writer
	c *Counter  // Counter for bytes written
}

// CountWriter creates an io.Writer that counts bytes written to w, using the provided [Counter].
func CountWriter(w io.Writer, c *Counter) io.Writer {
	return NewCounterWriter(w, c)
}

// NewCounterWriter creates a [CounterWriter] that counts bytes written to w, using the provided [Counter].
// If c is nil, a new Counter is created.
func NewCounterWriter(w io.Writer, c *Counter) *CounterWriter {
	if c == nil {
		c = new(Counter)
	}
	return &CounterWriter{w, c}
}

// Write writes to the underlying Writer and increments the counter by the number of bytes written.
// It returns the number of bytes written and any error encountered.
func (w *CounterWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	if n > 0 {
		w.c.Add(int64(n))
	}
	return
}

// Bytes returns the total number of bytes written.
func (w *CounterWriter) Bytes() int64 {
	return w.c.Get()
}
