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
	w io.Writer
	n uint64
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	if err != nil {
		return
	}

	atomic.AddUint64(&w.n, uint64(n))
	return
}

func (w *Writer) Count() uint64 {
	return atomic.LoadUint64(&w.n)
}

type WriteCloser struct {
	*Writer
	io.Closer
}

func NewWriterCloser(wc io.WriteCloser) *WriteCloser {
	return &WriteCloser{Writer: NewWriter(wc), Closer: wc}
}
