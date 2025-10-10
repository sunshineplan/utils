package counter

import (
	"io"
	"sync/atomic"
)

type Counter struct {
	n atomic.Int64
}

func (c *Counter) Add(delta int64) (new int64) {
	return c.n.Add(delta)
}

func (c *Counter) Get() int64 {
	return c.n.Load()
}

type CounterReader struct {
	r io.Reader
	c *Counter
}

func CountReader(r io.Reader, c *Counter) io.Reader {
	return NewCounterReader(r, c)
}

func NewCounterReader(r io.Reader, c *Counter) *CounterReader {
	if c == nil {
		c = new(Counter)
	}
	return &CounterReader{r, c}
}

func (r *CounterReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	if n > 0 {
		r.c.Add(int64(n))
	}
	return
}

func (r *CounterReader) Bytes() int64 {
	return r.c.Get()
}

type CounterWriter struct {
	w io.Writer
	c *Counter
}

func CountWriter(w io.Writer, c *Counter) io.Writer {
	return NewCounterWriter(w, c)
}

func NewCounterWriter(w io.Writer, c *Counter) *CounterWriter {
	if c == nil {
		c = new(Counter)
	}
	return &CounterWriter{w, c}
}

func (w *CounterWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	if n > 0 {
		w.c.Add(int64(n))
	}
	return
}

func (w *CounterWriter) Bytes() int64 {
	return w.c.Get()
}
