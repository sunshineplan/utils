package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"slices"

	"github.com/sunshineplan/utils/pool"
)

var utf8bom = []byte{0xEF, 0xBB, 0xBF}

// A Writer writes records using CSV encoding.
type Writer struct {
	*csv.Writer
	w             io.Writer
	utf8bom       bool
	fields        []field
	fieldsWritten bool
	zero          []string
	pool          *pool.Pool[[]string]
}

type field struct {
	name, tag string
}

// NewWriter returns a new Writer that writes to w.
func NewWriter(w io.Writer, utf8bom bool) *Writer {
	return &Writer{
		Writer:  csv.NewWriter(w),
		w:       w,
		utf8bom: utf8bom,
		pool:    pool.New[[]string](),
	}
}

// WriteFields writes fieldnames to w along with necessary utf8bom bytes. The fields must be a
// non-zero field struct or a non-zero length string slice, otherwise an error will be return.
// It can be run only once.
func (w *Writer) WriteFields(fields any) error {
	if w.fieldsWritten {
		return fmt.Errorf("fieldnames already be written")
	}
	switch f := fields.(type) {
	case []string:
		for _, i := range f {
			w.fields = append(w.fields, field{i, ""})
		}
	default:
		v := reflect.ValueOf(fields)
		if v.Kind() == reflect.Pointer {
			v = reflect.Indirect(v)
			if !v.IsValid() {
				return fmt.Errorf("can not get fieldnames from nil pointer struct")
			}
		}
		switch v.Kind() {
		case reflect.Struct:
			if v.NumField() == 0 {
				return fmt.Errorf("can not get fieldnames from zero field struct")
			}
			for i := range v.NumField() {
				if f := v.Type().Field(i); f.IsExported() {
					tag, _ := v.Type().Field(i).Tag.Lookup("csv")
					w.fields = append(w.fields, field{v.Type().Field(i).Name, tag})
				}
			}
		default:
			return fmt.Errorf("can not get fieldnames from fields which is not struct or string slice")
		}
	}
	if w.utf8bom {
		w.w.Write(utf8bom)
	}
	var record []string
	for _, i := range w.fields {
		if i.tag != "" {
			record = append(record, i.tag)
		} else {
			record = append(record, i.name)
		}
	}
	if err := w.Writer.Write(record); err != nil {
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	w.fieldsWritten = true
	w.zero = make([]string, len(w.fields))
	w.pool.New = func() *[]string {
		s := make([]string, len(w.fields))
		return &s
	}
	return nil
}

// Write writes a single CSV record to w along with any necessary quoting after fieldnames is written.
// A record is a map of strings or a struct. Writes are buffered, so Flush must eventually be called to
// ensure that the record is written to the underlying io.Writer.
func (w *Writer) Write(record any) error {
	if !w.fieldsWritten {
		return fmt.Errorf("fieldnames has not be written yet")
	}
	switch d := record.(type) {
	case []string:
		if len(d) == 0 {
			return nil
		}
		return w.Writer.Write(d)
	default:
		v := reflect.ValueOf(record)
		if v.Kind() == reflect.Interface {
			v = v.Elem()
		}
		if v.Kind() == reflect.Pointer {
			v = reflect.Indirect(v)
			if !v.IsValid() {
				return nil
			}
		}
		r := w.pool.Get()
		defer w.pool.Put(r)
		switch v.Kind() {
		case reflect.Map:
			if keyType := reflect.TypeOf(v.Interface()).Key(); keyType.Kind() == reflect.String {
				for i, field := range w.fields {
					key := reflect.Indirect(reflect.New(keyType))
					key.SetString(field.name)
					if v := v.MapIndex(key); v.IsValid() && v.Interface() != nil {
						(*r)[i], _ = marshalText(v.Interface())
					} else {
						(*r)[i] = ""
					}
				}
			} else {
				return fmt.Errorf("only can write record from map which is string kind")
			}
		case reflect.Struct:
			for i, field := range w.fields {
				var val reflect.Value
				var found bool
				for i := range v.NumField() {
					if tag, ok := v.Type().Field(i).Tag.Lookup("csv"); ok && tag == field.tag {
						val = v.FieldByName(field.name)
						found = true
						break
					}
				}
				if !found {
					if val = v.FieldByName(field.name); val.IsValid() && val.Interface() != nil {
						found = true
					}
				}
				if found {
					(*r)[i], _ = marshalText(val.Interface())
				} else {
					(*r)[i] = ""
				}
			}
		default:
			return fmt.Errorf("not support record format: %s", v.Kind())
		}
		if slices.Equal(*r, w.zero) {
			return nil
		}
		return w.Writer.Write(*r)
	}
}

// WriteAll writes multiple CSV records to w using Write and then calls Flush, returning any error from the Flush.
func (w *Writer) WriteAll(records any) error {
	switch s := records.(type) {
	case [][]string:
		for _, i := range s {
			if err := w.Write(i); err != nil {
				return err
			}
		}
	default:
		if reflect.TypeOf(records).Kind() != reflect.Slice {
			return fmt.Errorf("records is not slice")
		}
		v := reflect.ValueOf(records)
		for i := range v.Len() {
			if err := w.Write(v.Index(i).Interface()); err != nil {
				return err
			}
		}
	}
	w.Flush()
	return w.Error()
}
