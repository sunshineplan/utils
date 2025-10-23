package csv

import (
	"fmt"
	"io"
	"os"
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
	csvWriter := NewWriter(w, utf8bom)
	if len(fieldnames) == 0 {
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
