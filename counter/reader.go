package counter

import (
	"io"
	"sync/atomic"
)

type Reader interface {
	io.ReadCloser
	Count() int64
}

type reader struct {
	r io.Reader
	c io.Closer
	n atomic.Int64
}

func NewReader(r io.Reader) Reader {
	var c io.Closer
	if closer, ok := r.(io.Closer); ok {
		c = closer
	}
	reader := &reader{r: r, c: c}
	if _, ok := r.(io.WriterTo); ok {
		return readerWriterTo{reader}
	}
	return reader
}

func NewReaderCloser(rc io.ReadCloser) Reader {
	return NewReader(rc)
}

func (r *reader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	if err != nil {
		return
	}
	r.n.Add(int64(n))
	return
}

func (r *reader) Close() error {
	if r.c != nil {
		return r.c.Close()
	}
	return nil
}

func (r *reader) Count() int64 {
	return r.n.Load()
}

type readerWriterTo struct {
	*reader
}

func (r *readerWriterTo) WriteTo(w io.Writer) (n int64, err error) {
	return r.r.(io.WriterTo).WriteTo(w)
}
