package csv

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/sunshineplan/utils"
)

var utf8bom = []byte{0xEF, 0xBB, 0xBF}

// A Writer writes records using CSV encoding.
type Writer struct {
	*csv.Writer
	w             io.Writer
	utf8bom       bool
	fields        []field
	fieldsWritten bool
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
	}
}

// WriteFields writes fieldnames to w along with necessary utf8bom bytes. The fields must be a
// non-zero field struct or a non-zero length string slice, otherwise an error will be return.
// It can be run only once.
func (w *Writer) WriteFields(fields interface{}) error {
	if w.fieldsWritten {
		return fmt.Errorf("fieldnames already be written")
	}

	v := reflect.ValueOf(fields)

	switch v.Kind() {
	case reflect.Struct:
		if v.NumField() == 0 {
			return fmt.Errorf("can not get fieldnames from zero field struct")
		}

		for i := 0; i < v.NumField(); i++ {
			var f field
			if tag, ok := v.Type().Field(i).Tag.Lookup("csv"); ok {
				f = field{v.Type().Field(i).Name, tag}
			} else {
				f = field{v.Type().Field(i).Name, ""}
			}
			w.fields = append(w.fields, f)
		}
	case reflect.Slice:
		if v.Len() == 0 {
			return fmt.Errorf("can not get fieldnames from zero length slice")
		}

		fieldnames, ok := fields.([]string)
		if !ok {
			return fmt.Errorf("only can get fieldnames from slice which is string slice")
		}
		for _, i := range fieldnames {
			w.fields = append(w.fields, field{i, ""})
		}
	default:
		return fmt.Errorf("can not get fieldnames from fields which is not struct or string slice")
	}

	w.fields = utils.Deduplicate(w.fields).([]field)

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
			for i, field := range w.fields {
				if v := v.MapIndex(reflect.ValueOf(field.name)); v.IsValid() && v.Interface() != nil {
					if vi := v.Interface(); reflect.TypeOf(vi).Kind() == reflect.String {
						r[i] = vi.(string)
					} else {
						b, _ := json.Marshal(vi)
						r[i] = string(b)
					}
				}
			}
		} else {
			return fmt.Errorf("only can write record from map which is string")
		}
	case reflect.Struct:
		for i, field := range w.fields {
			var val reflect.Value
			var found bool
			for i := 0; i < v.NumField(); i++ {
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
				if v := val.Interface(); reflect.TypeOf(v).Kind() == reflect.String {
					r[i] = v.(string)
				} else {
					b, _ := json.Marshal(v)
					r[i] = string(b)
				}
			}
		}
	default:
		return fmt.Errorf("not support record format: %s", v.Kind())
	}

	return w.Writer.Write(r)
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
	w.Flush()

	return w.Error()
}
