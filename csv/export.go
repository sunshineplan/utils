package csv

import (
	"fmt"
	"io"
	"os"
	"reflect"
)

// Export writes slice as csv format with fieldnames to writer w.
func Export(fieldnames []string, slice interface{}, w io.Writer) error {
	return export(fieldnames, slice, w, false)
}

// ExportFile writes slice as csv format with fieldnames to file.
func ExportFile(fieldnames []string, slice interface{}, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	return export(fieldnames, slice, f, false)
}

// ExportUTF8 writes slice as utf8 csv format with fieldnames to writer w.
func ExportUTF8(fieldnames []string, slice interface{}, w io.Writer) error {
	return export(fieldnames, slice, w, true)
}

// ExportUTF8File writes slice as utf8 csv format with fieldnames to file.
func ExportUTF8File(fieldnames []string, slice interface{}, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	return export(fieldnames, slice, f, true)
}

func export(fieldnames []string, slice interface{}, w io.Writer, utf8bom bool) (err error) {
	if reflect.TypeOf(slice).Kind() != reflect.Slice {
		return fmt.Errorf("rows is not slice")
	}

	csvWriter := NewWriter(w, utf8bom)

	rows := reflect.ValueOf(slice)
	if fieldnames == nil {
		if rows.Len() == 0 {
			return fmt.Errorf("can't get struct fieldnames from zero length slice")
		}

		err = csvWriter.WriteFields(rows.Index(0).Interface())
	} else {
		err = csvWriter.WriteFields(fieldnames)
	}
	if err != nil {
		return
	}

	return csvWriter.WriteAll(slice)
}
