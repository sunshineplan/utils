package csv

import (
	"fmt"
	"io"
	"os"
)

// Export writes slice as csv format with fieldnames to writer w.
func Export[T any](fieldnames []string, slice []T, w io.Writer) error {
	return export(fieldnames, slice, w, false)
}

// ExportFile writes slice as csv format with fieldnames to file.
func ExportFile[T any](fieldnames []string, slice []T, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	return export(fieldnames, slice, f, false)
}

// ExportUTF8 writes slice as utf8 csv format with fieldnames to writer w.
func ExportUTF8[T any](fieldnames []string, slice []T, w io.Writer) error {
	return export(fieldnames, slice, w, true)
}

// ExportUTF8File writes slice as utf8 csv format with fieldnames to file.
func ExportUTF8File[T any](fieldnames []string, slice []T, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	return export(fieldnames, slice, f, true)
}

func export[T any](fieldnames []string, slice []T, w io.Writer, utf8bom bool) (err error) {
	csvWriter := NewWriter(w, utf8bom)

	if fieldnames == nil {
		if len(slice) == 0 {
			return fmt.Errorf("can't get struct fieldnames from zero length slice")
		}

		err = csvWriter.WriteFields(slice[0])
	} else {
		err = csvWriter.WriteFields(fieldnames)
	}
	if err != nil {
		return
	}

	return csvWriter.WriteAll(slice)
}
