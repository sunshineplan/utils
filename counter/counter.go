package counter

import (
	"io"
	"sync/atomic"
)

type Counter atomic.Int64

func (c *Counter) Add(delta int64) (new int64) {
	return (*atomic.Int64)(c).Add(delta)
}

func (c *Counter) Load() int64 {
	return (*atomic.Int64)(c).Load()
}

func (c *Counter) AddWriter(w io.Writer) io.Writer {
	return newWriter(c, w)
}

func (c *Counter) AddReader(r io.Reader) io.Reader {
	return newReader(c, r)
}
