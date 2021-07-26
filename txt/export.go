package txt

import (
	"io"
	"os"
)

// Export writes content to writer w.
func Export(content []string, w io.Writer) error {
	for _, i := range content {
		if _, err := io.WriteString(w, i+"\n"); err != nil {
			return err
		}
	}

	return nil
}

// ExportFile writes content to file.
func ExportFile(content []string, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	return Export(content, f)
}
