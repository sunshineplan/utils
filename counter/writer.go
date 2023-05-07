package counter

import (
	"io"
	"sync/atomic"
)

type Writer interface {
	io.WriteCloser
	Count() int64
}

type writer struct {
	w io.Writer
	c io.Closer
	n atomic.Int64
}

func NewWriter(w io.Writer) Writer {
	var c io.Closer
	if closer, ok := w.(io.Closer); ok {
		c = closer
	}
	writer := &writer{w: w, c: c}
	if _, ok := w.(io.ReaderFrom); ok {
		return writerReaderFrom{writer}
	}
	return writer
}

func NewWriterCloser(wc io.WriteCloser) Writer {
	return NewWriter(wc)
}

func (w *writer) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	if err != nil {
		return
	}
	w.n.Add(int64(n))
	return
}

func (w *writer) Close() error {
	if w.c != nil {
		return w.c.Close()
	}
	return nil
}

func (w *writer) Count() int64 {
	return w.n.Load()
}

type writerReaderFrom struct {
	*writer
}

func (w *writerReaderFrom) ReadFrom(r Reader) (n int64, err error) {
	return w.w.(io.ReaderFrom).ReadFrom(r)
}
