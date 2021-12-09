package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// A Reader reads records from a CSV-encoded file.
type Reader struct {
	*csv.Reader
	next    []string
	nextErr error
}

// NewReader returns a new Reader that reads from r.
func NewReader(r io.Reader) *Reader {
	return &Reader{Reader: csv.NewReader(r)}
}

// Next prepares the next result row for reading with the Scan method.
func (r *Reader) Next() bool {
	r.next, r.nextErr = r.Read()
	return r.nextErr == nil
}

// Scan copies the columns in the current record into the values pointed at by dest.
// The number of values in dest must be the same as the number of columns in record.
func (r *Reader) Scan(dest ...interface{}) error {
	if r.next == nil && r.nextErr == nil {
		return fmt.Errorf("Scan called without calling Next")
	}

	if r.nextErr != nil {
		return r.nextErr
	}
	if len(dest) != len(r.next) {
		return fmt.Errorf("expected %d destination arguments in Scan, not %d", len(r.next), len(dest))
	}

	for i, v := range r.next {
		if err := convertAssign(dest[i], v); err != nil {
			return fmt.Errorf("Scan error on field index %d: %v", i, err)
		}
	}

	return nil
}

// Rows is the records of a csv file. Its cursor starts before
// the first row of the result set. Use Next to advance from row to row.
type Rows struct {
	*Reader
	io.Closer
}

// FromReader returns Rows reads from r.
func FromReader(r io.Reader) *Rows {
	if closer, ok := r.(io.Closer); ok {
		return &Rows{NewReader(r), closer}
	}
	return &Rows{NewReader(r), io.NopCloser(r)}
}

// ReadFile returns Rows reads from file.
func ReadFile(file string) (*Rows, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	return FromReader(f), nil
}

// Next prepares the next result row for reading with the Scan method.
func (rs *Rows) Next() bool {
	return rs.Reader.Next()
}

// Scan copies the columns in the current row into the values pointed at by dest.
// The number of values in dest must be the same as the number of columns in Rows.
func (rs *Rows) Scan(dest ...interface{}) error {
	return rs.Reader.Scan(dest...)
}
