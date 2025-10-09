package counter

import "io"

type readerCounter struct {
	*Counter
	r io.Reader
}

func newReaderCounter(n *Counter, r io.Reader) io.Reader {
	reader := &readerCounter{n, r}
	if _, ok := r.(io.WriterTo); ok {
		return writerTo{reader}
	}
	return reader
}

func (r *readerCounter) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	if n > 0 {
		r.Add(int64(n))
	}
	return
}

type writerTo struct {
	*readerCounter
}

func (r *writerTo) WriteTo(w io.Writer) (n int64, err error) {
	n, err = io.Copy(w, r.r)
	if n > 0 {
		r.Add(n)
	}
	return
}

type writerCounter struct {
	*Counter
	w io.Writer
}

func newWriterCounter(n *Counter, w io.Writer) io.Writer {
	writer := &writerCounter{n, w}
	if _, ok := w.(io.ReaderFrom); ok {
		return readerFrom{writer}
	}
	return writer
}

func (w *writerCounter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	if n > 0 {
		w.Add(int64(n))
	}
	return
}

type readerFrom struct {
	*writerCounter
}

func (w *readerFrom) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = io.Copy(w.w, r)
	if n > 0 {
		w.Add(n)
	}
	return
}
