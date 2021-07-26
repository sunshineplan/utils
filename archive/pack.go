package archive

import (
	"errors"
	"io"
)

// Pack creates an archive from File struct.
func Pack(w io.Writer, format Format, files ...File) error {
	switch format {
	case ZIP:
		return packZip(w, files...)
	case TAR:
		return packTar(w, files...)
	default:
		return errors.New("unknow format")
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
