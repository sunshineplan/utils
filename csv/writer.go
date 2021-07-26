package csv

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

var utf8bom = []byte{0xEF, 0xBB, 0xBF}

// A Writer writes records using CSV encoding.
type Writer struct {
	writer        io.Writer
	csvWriter     *csv.Writer
	utf8bom       bool
	fields        []string
	fieldsWritten bool
}

// NewWriter returns a new Writer that writes to w.
func NewWriter(w io.Writer, utf8bom bool) *Writer {
	return &Writer{
		writer:    w,
		csvWriter: csv.NewWriter(w),
		utf8bom:   utf8bom,
	}
}

// WriteFields writes fieldnames to w along with necessary utf8bom bytes.
// It can be run only once.
func (w *Writer) WriteFields(fields interface{}) error {
	if w.fieldsWritten {
		return fmt.Errorf("fieldnames already be written")
	}

	var fieldnames []string
	v := reflect.ValueOf(fields)
	switch v.Kind() {
	case reflect.Struct:
		if v.NumField() == 0 {
			return fmt.Errorf("can not get fieldnames from zero field struct")
		}

		for i := 0; i < v.NumField(); i++ {
			fieldnames = append(fieldnames, v.Type().Field(i).Name)
		}
	case reflect.Slice:
		if v.Len() == 0 {
			return fmt.Errorf("can not get fieldnames from zero length slice")
		}

		var ok bool
		if fieldnames, ok = fields.([]string); !ok {
			return fmt.Errorf("only can get fieldnames from slice which is string slice")
		}
	default:
		return fmt.Errorf("can not get fieldnames from interface which is not struct or string slice")
	}

	w.fields = fieldnames

	if w.utf8bom {
		w.writer.Write(utf8bom)
	}

	if err := w.csvWriter.Write(fieldnames); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}

	w.fieldsWritten = true

	return nil
}

// Write writes a single CSV record to w along with any necessary quoting after fieldnames is written.
// A record is a map of strings or a struct. Writes are buffered, so Flush must eventually be called to
// ensure that the record is written to the underlying io.Writer.
func (w *Writer) Write(record interface{}) error {
	if !w.fieldsWritten {
		return fmt.Errorf("fieldnames has not be written yet")
	}

	v := reflect.ValueOf(record)
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	r := make([]string, len(w.fields))
	switch v.Kind() {
	case reflect.Map:
		if reflect.TypeOf(v.Interface()).Key().Name() == "string" {
			for index, fieldname := range w.fields {
				if v := v.MapIndex(reflect.ValueOf(fieldname)); v.IsValid() && v.Interface() != nil {
					if vi := v.Interface(); reflect.TypeOf(vi).Kind() == reflect.String {
						r[index] = vi.(string)
					} else {
						b, _ := json.Marshal(vi)
						r[index] = string(b)
					}
				}
			}
		} else {
			return fmt.Errorf("only can write record from map which is string")
		}
	case reflect.Struct:
		for index, fieldname := range w.fields {
			if v := v.FieldByName(fieldname); v.IsValid() && v.Interface() != nil {
				if vi := v.Interface(); reflect.TypeOf(vi).Kind() == reflect.String {
					r[index] = vi.(string)
				} else {
					b, _ := json.Marshal(vi)
					r[index] = string(b)
				}
			}
		}
	default:
		return fmt.Errorf("not support record format: %s", v.Kind())
	}

	return w.csvWriter.Write(r)
}

// WriteAll writes multiple CSV records to w using Write and then calls Flush, returning any error from the Flush.
func (w *Writer) WriteAll(records interface{}) error {
	if reflect.TypeOf(records).Kind() != reflect.Slice {
		return fmt.Errorf("records is not slice")
	}

	v := reflect.ValueOf(records)
	for i := 0; i < v.Len(); i++ {
		if err := w.Write(v.Index(i).Interface()); err != nil {
			return err
		}
	}

	return w.Flush()
}

// Error reports any error that has occurred during a previous Write or Flush.
func (w *Writer) Error() error {
	return w.csvWriter.Error()
}

// Flush writes any buffered data to the underlying io.Writer.
func (w *Writer) Flush() error {
	w.csvWriter.Flush()

	return w.Error()
}
