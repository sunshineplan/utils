package txt

import (
	"bufio"
	"io"
)

// Writer implements buffering for an io.Writer object. If an error occurs writing to a Writer,
// no more data will be accepted and all subsequent writes, and Flush, will return the error.
// After all data has been written, the client should call the Flush method to guarantee all data
// has been forwarded to the underlying io.Writer.
type Writer struct {
	*bufio.Writer
	UseCRLF bool // True to use \r\n as the line terminator
}

// NewWriter returns a new text Writer whose buffer has the default size.
func NewWriter(w io.Writer) *Writer {
	return &Writer{Writer: bufio.NewWriter(w)}
}

// NewWriterSize returns a new text Writer whose buffer has at least the specified size.
// If the argument io.Writer is already a Writer with large enough size, it returns the underlying Writer.
func NewWriterSize(w io.Writer, size int) *Writer {
	return &Writer{Writer: bufio.NewWriterSize(w, size)}
}

// WriteLine writes a string end with the line terminator. It returns the number of bytes written.
// If the count is less than expected, it also returns an error explaining why the write is short.
func (w *Writer) WriteLine(s string) (int, error) {
	if w.UseCRLF {
		return w.WriteString(s + "\r\n")
	}

	return w.WriteString(s + "\n")
}

// WriteAll writes contents to w using WriteLine and then calls Flush,
// returning any error from the Flush.
func (w *Writer) WriteAll(contents []string) error {
	for _, i := range contents {
		if _, err := w.WriteLine(i); err != nil {
			return err
		}
	}

	return w.Flush()
}
