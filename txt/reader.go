package txt

import (
	"bufio"
	"io"
	"iter"
	"os"
)

// Reader provides buffered reading from an io.Reader, splitting input by lines.
type Reader struct {
	scanner *bufio.Scanner
}

// NewReader returns a new Reader that reads from r using a buffered scanner.
// It splits input by lines using the default bufio.Scanner line-splitting behavior (\n).
func NewReader(r io.Reader) *Reader {
	return &Reader{bufio.NewScanner(r)}
}

// Iter returns an iterator over the lines read from the underlying io.Reader.
// Each iteration yields a line and a nil error, or an empty string and an error
// if the scanner encounters an error.
func (r *Reader) Iter() iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for r.scanner.Scan() {
			if !yield(r.scanner.Text(), nil) {
				return
			}
		}
		if err := r.scanner.Err(); err != nil {
			yield("", err)
		}
	}
}

// ReadAll reads all lines from r and returns them as a slice of strings.
// It returns an error if the underlying scanner encounters an error.
func ReadAll(r io.Reader) ([]string, error) {
	var s []string
	for i, err := range NewReader(r).Iter() {
		if err != nil {
			return nil, err
		}
		s = append(s, i)
	}
	return s, nil
}

// ReadFile reads all lines from the specified file and returns them as a slice of strings.
// It returns an error if the file cannot be opened or read.
func ReadFile(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadAll(f)
}
