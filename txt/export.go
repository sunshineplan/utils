package txt

import (
	"io"
	"os"
)

// Export writes the contents to w using a buffered Writer.
// It returns any error encountered during writing or flushing.
func Export(contents []string, w io.Writer) error {
	return NewWriter(w).WriteAll(contents)
}

// ExportFile writes the contents to the specified file, overwriting it if it exists.
// The file is created with default permissions (0666, subject to umask).
// It returns any error encountered during file creation or writing.
func ExportFile(contents []string, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return Export(contents, f)
}
