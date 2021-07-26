package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// Rows is the records of a csv file. Its cursor starts before
// the first row of the result set. Use Next to advance from row to row.
type Rows struct {
	fields   []string
	records  chan []string
	lastcols []string
	closed   bool
}

func lineCounter(r io.Reader) (count int, err error) {
	buf := make([]byte, 32*1024)

	for {
		var c int
		c, err = r.Read(buf)
		if err == io.EOF {
			err = nil
			return
		}
		if err != nil {
			return
		}
		count += bytes.Count(buf[:c], []byte{'\n'})
	}
}

// ReadAll reads all the records from r.
func ReadAll(r io.Reader) (*Rows, error) {
	var buf bytes.Buffer
	count, err := lineCounter(io.TeeReader(r, &buf))
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(&buf)
	rs := &Rows{records: make(chan []string, count+1)}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rs.records <- record
	}

	switch len(rs.records) {
	case 0:
		return nil, fmt.Errorf("empty csv file")
	case 1:
		rs.fields = <-rs.records
		close(rs.records)
		rs.closed = true
	default:
		rs.fields = <-rs.records
	}

	return rs, nil
}

// ReadFile reads all records from file.
func ReadFile(file string) (*Rows, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	return ReadAll(f)
}

// Fields returns the fieldnames.
func (rs *Rows) Fields() []string {
	return rs.fields
}

// Next prepares the next result row for reading with the Scan method.
func (rs *Rows) Next() bool {
	if rs.closed {
		return false
	}

	if len(rs.records) > 0 {
		rs.lastcols = <-rs.records
		return true
	}

	close(rs.records)
	rs.closed = true

	return false
}

// Scan copies the columns in the current row into the values pointed at by dest.
// The number of values in dest must be the same as the number of columns in Rows.
func (rs *Rows) Scan(dest ...interface{}) error {
	if rs.closed {
		return fmt.Errorf("Rows are closed")
	}

	if len(dest) != len(rs.fields) {
		return fmt.Errorf("expected %d destination arguments in Scan, not %d", len(rs.fields), len(dest))
	}

	if rs.lastcols == nil {
		return fmt.Errorf("Scan called without calling Next")
	}

	for i, v := range rs.lastcols {
		if err := convertAssign(dest[i], v); err != nil {
			return fmt.Errorf("Scan error on field index %d, name %q: %v", i, rs.Fields()[i], err)
		}
	}

	return nil
}
