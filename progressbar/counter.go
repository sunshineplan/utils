package progressbar

import (
	"io"

	"github.com/sunshineplan/utils/counter"
)

type genericCounter interface {
	Add(int64) int64
	Write([]byte) (int, error)
	Get() int64
}

var (
	_ genericCounter = new(numberCounter)
	_ genericCounter = new(writerCounter)
)

func newNumberCounter() genericCounter {
	return new(numberCounter)
}

func newWriterCounter(w io.Writer) genericCounter {
	return &writerCounter{counter.NewCounterWriter(w, new(counter.Counter))}
}

type numberCounter struct {
	counter.Counter
}

func (c *numberCounter) Write(_ []byte) (n int, err error) { return }

type writerCounter struct {
	*counter.CounterWriter
}

func (c *writerCounter) Add(_ int64) (n int64) { return }
func (c *writerCounter) Get() int64            { return c.CounterWriter.Bytes() }
