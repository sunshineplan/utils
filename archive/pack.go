package archive

import (
	"io"
	"os"
)

// Pack creates an archive from File struct.
func Pack(w io.Writer, format Format, files ...File) error {
	switch format {
	case ZIP:
		return packZip(w, files...)
	case TAR:
		return packTar(w, files...)
	default:
		return ErrFormat
	}
}

// PackFromFiles creates an archive from files.
func PackFromFiles(w io.Writer, format Format, files ...string) error {
	fs, err := readFiles(files...)
	if err != nil {
		return err
	}
	return Pack(w, format, fs...)
}

func readFiles(files ...string) (fs []File, err error) {
	for _, f := range files {
		var file File
		file.Name = f
		file.Body, err = os.ReadFile(f)
		if err != nil {
			return
		}
		fs = append(fs, file)
	}
	return
}
