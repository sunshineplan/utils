package counter

import "io"

type reader struct {
	*Counter
	r io.Reader
}

func newReader(n *Counter, r io.Reader) io.Reader {
	reader := &reader{n, r}
	if _, ok := r.(io.WriterTo); ok {
		return readerWriterTo{reader}
	}
	return reader
}

func (r *reader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	if n > 0 {
		r.Add(int64(n))
	}
	return
}

type readerWriterTo struct {
	*reader
}

func (r *readerWriterTo) WriteTo(w io.Writer) (n int64, err error) {
	n, err = r.r.(io.WriterTo).WriteTo(w)
	if n > 0 {
		r.Add(n)
	}
	return
}

type writer struct {
	*Counter
	w io.Writer
}

func newWriter(n *Counter, w io.Writer) io.Writer {
	writer := &writer{n, w}
	if _, ok := w.(io.ReaderFrom); ok {
		return writerReaderFrom{writer}
	}
	return writer
}

func (w *writer) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	if n > 0 {
		w.Add(int64(n))
	}
	return
}

type writerReaderFrom struct {
	*writer
}

func (w *writerReaderFrom) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = w.w.(io.ReaderFrom).ReadFrom(r)
	if n > 0 {
		w.Add(n)
	}
	return
}
