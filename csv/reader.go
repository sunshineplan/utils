package csv

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// A Reader reads records from a CSV-encoded file.
type Reader struct {
	*csv.Reader
	io.Closer

	fields []string

	next    []string
	nextErr error
}

// NewReader returns a new Reader that reads from r.
func NewReader(r io.Reader, hasFields bool) *Reader {
	var reader *Reader
	if closer, ok := r.(io.Closer); ok {
		reader = &Reader{Reader: csv.NewReader(r), Closer: closer}
	}
	reader = &Reader{Reader: csv.NewReader(r), Closer: io.NopCloser(r)}

	if hasFields {
		var err error
		reader.fields, err = reader.Read()
		if err != nil {
			panic(err)
		}
	}
	return reader
}

// ReadFile returns Reader reads from file.
func ReadFile(file string, hasFields bool) (*Reader, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	return NewReader(f, hasFields), nil
}

// SetFields sets csv fields.
func (r *Reader) SetFields(fields []string) {
	r.fields = fields
}

// Next prepares the next record for reading with the Scan or Decode method.
func (r *Reader) Next() bool {
	r.next, r.nextErr = r.Read()
	return r.nextErr != io.EOF
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

// Decode will unmarshal the current record into dest.
// If column's value is like "[...]", it will be treated as slice.
func (r *Reader) Decode(dest interface{}) error {
	if len(r.fields) == 0 {
		return fmt.Errorf("csv fields is not parsed")
	}

	if r.next == nil && r.nextErr == nil {
		return fmt.Errorf("Decode called without calling Next")
	}

	if r.nextErr != nil {
		return r.nextErr
	}

	m := make(map[string]interface{})
	for i, field := range r.fields {
		if len(r.next) > i {
			m[field] = convert(r.next[i])
		}
	}

	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, dest)
}

// DecodeAll decodes each record from r into dest.
func DecodeAll[T any](r io.Reader, dest *[]T) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	reader := NewReader(r, true)
	defer reader.Close()

	var res []T
	for reader.Next() {
		var t T
		if err = reader.Decode(&t); err != nil {
			return
		}
		res = append(res, t)
	}
	*dest = res

	return
}

// DecodeFile decodes each record from file into dest.
func DecodeFile[T any](file string, dest *[]T) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	return DecodeAll(f, dest)
}
