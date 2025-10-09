package counter

import (
	"io"
	"sync/atomic"
	"time"
)

type Counter struct {
	v atomic.Int64
}

func (c *Counter) Add(delta int64) (new int64) {
	return c.v.Add(delta)
}

func (c *Counter) Load() int64 {
	return c.v.Load()
}

func (c *Counter) AddWriter(w io.Writer) io.Writer {
	return newWriter(c, w)
}

func (c *Counter) AddReader(r io.Reader) io.Reader {
	return newReader(c, r)
}

type RateCounter struct {
	Counter
	start time.Time
}

func NewRateCounter() *RateCounter {
	return &RateCounter{start: time.Now()}
}

func (rc *RateCounter) Reset() {
	rc.Counter.v.Store(0)
	rc.start = time.Now()
}

func (rc *RateCounter) Rate() float64 {
	duration := time.Since(rc.start).Seconds()
	if duration == 0 {
		return 0
	}
	return float64(rc.Load()) / duration
}
