package csv

import (
	"io"
	"os"
	"reflect"
)

// Export writes slice as csv format with fieldnames to writer w.
func Export[S ~[]E, E any](fieldnames []string, slice S, w io.Writer) error {
	return export(fieldnames, slice, w, false)
}

// ExportFile writes slice as csv format with fieldnames to file.
func ExportFile[S ~[]E, E any](fieldnames []string, slice S, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return export(fieldnames, slice, f, false)
}

// ExportUTF8 writes slice as utf8 csv format with fieldnames to writer w.
func ExportUTF8[S ~[]E, E any](fieldnames []string, slice S, w io.Writer) error {
	return export(fieldnames, slice, w, true)
}

// ExportUTF8File writes slice as utf8 csv format with fieldnames to file.
func ExportUTF8File[S ~[]E, E any](fieldnames []string, slice S, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return export(fieldnames, slice, f, true)
}

func export[S ~[]E, E any](fieldnames []string, slice S, w io.Writer, utf8bom bool) (err error) {
	var fields any
	if len(fieldnames) == 0 {
		t := reflect.TypeFor[E]()
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if kind := t.Kind(); kind == reflect.Struct || kind == reflect.Map {
			fields = reflect.Zero(t).Interface()
		}
	} else {
		fields = fieldnames
	}
	csvWriter := NewWriter(w, utf8bom)
	if fields != nil {
		if err = csvWriter.WriteFields(fields); err != nil {
			return
		}
	} else {
		csvWriter.fieldsWritten = true
		csvWriter.zero = make([]string, 1)
		csvWriter.pool.New = func() *[]string {
			s := make([]string, 1)
			return &s
		}
	}
	return csvWriter.WriteAll(slice)
}
