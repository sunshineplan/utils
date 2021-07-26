package archive

import (
	"archive/zip"
	"bytes"
	"io"
	"log"
)

const zipMagic = "PK\x03\x04"

func packZip(w io.Writer, files ...File) error {
	zw := zip.NewWriter(w)

	for _, file := range files {
		f, err := zw.Create(file.Name)
		if err != nil {
			return err
		}
		if _, err := f.Write(file.Body); err != nil {
			return err
		}
	}

	return zw.Close()
}

func unpackZip(b []byte) ([]File, error) {
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}

	var fs []File
	for _, f := range r.File {
		switch {
		case f.FileInfo().IsDir():
			fs = append(fs, File{Name: f.Name, IsDir: true})
		case f.FileInfo().Mode().IsRegular():
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, rc); err != nil {
				return nil, err
			}
			fs = append(fs, File{Name: f.Name, Body: buf.Bytes()})
			if err := rc.Close(); err != nil {
				return nil, err
			}
		default:
			log.Printf(
				"ExtractZip: uknown type: %d in %s",
				f.FileInfo().Mode(),
				f.Name)
		}
	}

	return fs, nil
}
