package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"log"
)

const tarMagic = "\x1f\x8b\x08\x00"

func packTar(w io.Writer, files ...File) error {
	gw := gzip.NewWriter(w)
	tw := tar.NewWriter(gw)

	for _, file := range files {
		header := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if _, err := tw.Write(file.Body); err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return err
	}

	return gw.Close()
}

func unpackTar(b []byte) ([]File, error) {
	gr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gr)

	var fs []File
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			fs = append(fs, File{Name: header.Name, IsDir: true})
		case tar.TypeReg:
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, tr); err != nil {
				return nil, err
			}
			fs = append(fs, File{Name: header.Name, Body: buf.Bytes()})
		default:
			log.Printf(
				"ExtractTarGz: uknown type: %v in %s",
				header.Typeflag,
				header.Name)
		}
	}

	return fs, nil
}
