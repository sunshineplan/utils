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
	n uint64
}

func NewReader(r io.Reader) *Reader {
	return &Reader{Reader: r}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if err != nil {
		return
	}

	atomic.AddUint64(&r.n, uint64(n))
	return
}

func (r *Reader) Count() uint64 {
	return atomic.LoadUint64(&r.n)
}

type ReadCloser struct {
	*Reader
	io.Closer
}

func NewReaderCloser(rc io.ReadCloser) *ReadCloser {
	return &ReadCloser{Reader: NewReader(rc), Closer: rc}
}
