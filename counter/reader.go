package counter

import (
	"io"
	"sync/atomic"
)

var (
	_ io.Reader     = &Reader{}
	_ io.ReadCloser = &ReadCloser{}
)

type Reader struct {
	io.Reader
	n atomic.Int64
}

func NewReader(r io.Reader) *Reader {
	return &Reader{Reader: r}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if err != nil {
		return
	}
	r.n.Add(int64(n))
	return
}

func (r *Reader) Count() int64 {
	return r.n.Load()
}

type ReadCloser struct {
	*Reader
	io.Closer
}

func NewReaderCloser(rc io.ReadCloser) *ReadCloser {
	return &ReadCloser{Reader: NewReader(rc), Closer: rc}
}
