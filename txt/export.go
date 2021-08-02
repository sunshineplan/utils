package txt

import (
	"io"
	"os"
)

// Export writes contents to writer w.
func Export(contents []string, w io.Writer) error {
	return NewWriter(w).WriteAll(contents)
}

// ExportFile writes contents to file.
func ExportFile(contents []string, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	return Export(contents, f)
}
