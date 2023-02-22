package counter

import (
	"io"
	"sync/atomic"
)

var (
	_ io.Writer      = &Writer{}
	_ io.WriteCloser = &WriteCloser{}
)

type Writer struct {
	io.Writer
	n atomic.Int64
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{Writer: w}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	if err != nil {
		return
	}
	w.n.Add(int64(n))
	return
}

func (w *Writer) Count() int64 {
	return w.n.Load()
}

type WriteCloser struct {
	*Writer
	io.Closer
}

func NewWriterCloser(wc io.WriteCloser) *WriteCloser {
	return &WriteCloser{Writer: NewWriter(wc), Closer: wc}
}
