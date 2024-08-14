package txt

import (
	"bufio"
	"io"
	"iter"
	"os"
)

type Reader struct {
	scanner *bufio.Scanner
}

func NewReader(r io.Reader) *Reader {
	return &Reader{bufio.NewScanner(r)}
}

func (r *Reader) Iter() iter.Seq[string] {
	return func(yield func(string) bool) {
		for r.scanner.Scan() {
			if !yield(r.scanner.Text()) {
				return
			}
		}
	}
}

// ReadAll reads all contents from r.
func ReadAll(r io.Reader) []string {
	var s []string
	for i := range NewReader(r).Iter() {
		s = append(s, i)
	}
	return s
}

// ReadFile reads all contents from file.
func ReadFile(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadAll(f), nil
}
