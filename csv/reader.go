package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
)

// A Reader reads records from a CSV-encoded file.
type Reader struct {
	*csv.Reader
	closer io.Closer

	once      sync.Once
	fields    []string
	hasFields bool

	next    []string
	nextErr error
}

// NewReader returns a new Reader that reads from r.
func NewReader(r io.Reader, hasFields bool) (*Reader, error) {
	reader := &Reader{Reader: csv.NewReader(r)}
	if closer, ok := r.(io.Closer); ok {
		reader.closer = closer
	}
	if hasFields {
		var err error
		reader.fields, err = reader.Read()
		if err != nil {
			return nil, err
		}
		reader.hasFields = true
	}
	return reader, nil
}

// ReadFile returns Reader reads from file.
func ReadFile(file string, hasFields bool) (*Reader, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	reader, err := NewReader(f, hasFields)
	if err != nil {
		f.Close()
		return nil, err
	}
	return reader, nil
}

// Read reads one record (a slice of fields) from r.
// If the record has an unexpected number of fields,
// Read returns the record along with the error [ErrFieldCount].
// If the record contains a field that cannot be parsed,
// Read returns a partial record along with the parse error.
// The partial record contains all fields read before the error.
// If there is no data left to be read, Read returns nil, [io.EOF].
// If [Reader.ReuseRecord] is true, the returned slice may be shared
// between multiple calls to Read.
func (r *Reader) Read() (record []string, err error) {
	record, err = r.Reader.Read()
	if err == nil {
		r.once.Do(func() {
			if len(record) > 0 {
				record[0] = strings.TrimPrefix(record[0], string(utf8bom))
			}
		})
	}
	return
}

// SetFields sets csv fields.
func (r *Reader) SetFields(fields []string) {
	r.fields = fields
	r.hasFields = true
}

// Next prepares the next record for reading with the Scan or Decode method.
func (r *Reader) Next() bool {
	r.next, r.nextErr = r.Read()
	return r.nextErr != io.EOF
}

// Scan copies the columns in the current record into the values pointed at by dest.
// The number of values in dest must be the same as the number of columns in record.
func (r *Reader) Scan(dest ...any) error {
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
		if err := setCell(dest[i], v); err != nil {
			return fmt.Errorf("Scan error on field index %d: %v", i, err)
		}
	}
	return nil
}

// Decode will unmarshal the current record into dest.
// If column's value is like "[...]", it will be treated as slice.
func (r *Reader) Decode(dest any) error {
	if r.hasFields {
		if len(r.fields) == 0 {
			return fmt.Errorf("csv fields is not parsed")
		}
		if r.next == nil && r.nextErr == nil {
			return fmt.Errorf("Decode called without calling Next")
		}
		if r.nextErr != nil {
			return r.nextErr
		}
		m := make(map[string]string)
		for i, field := range r.fields {
			if len(r.next) > i {
				m[field] = r.next[i]
			}
		}
		return setRow(dest, m)
	}
	return setCell(dest, r.next[0])
}

// Close closes the underlying reader if it implements the io.Closer interface.
func (r *Reader) Close() error {
	if r.closer != nil {
		return r.closer.Close()
	}
	return nil
}

// DecodeAll decodes each record from r into dest.
func DecodeAll[S ~[]E, E any](r io.Reader, dest *S) (err error) {
	t := reflect.TypeFor[E]()
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	var reader *Reader
	if kind := t.Kind(); kind == reflect.Struct || kind == reflect.Map {
		reader, err = NewReader(r, true)
	} else {
		reader, err = NewReader(r, false)
	}
	if err != nil {
		return
	}
	*dest = nil
	for reader.Next() {
		var t E
		if err = reader.Decode(&t); err != nil {
			*dest = nil
			return
		}
		*dest = append(*dest, t)
	}
	return
}

// DecodeFile decodes each record from file into dest.
func DecodeFile[S ~[]E, E any](file string, dest *S) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return DecodeAll(f, dest)
}
