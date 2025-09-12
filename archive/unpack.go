package archive

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Unpack decompresses an archive to File struct.
func Unpack(r io.Reader) ([]File, error) {
	rr := asReader(r)
	_, format := isArchive(rr)
	switch format {
	case ZIP:
		b, err := io.ReadAll(rr)
		if err != nil {
			return nil, err
		}
		return unpackZip(bytes.NewReader(b), int64(len(b)))
	case TAR:
		return unpackTar(rr)
	default:
		return nil, ErrFormat
	}
}

// UnpackToFiles decompresses an archive to files.
func UnpackToFiles(r io.Reader, dest string) error {
	files, err := Unpack(r)
	if err != nil {
		return err
	}

	for _, file := range files {
		fpath := filepath.Join(dest, file.Name)
		if file.IsDir {
			dir, err := os.Stat(fpath)
			if err != nil {
				if os.IsNotExist(err) {
					if err := os.MkdirAll(fpath, 0755); err != nil {
						return err
					}
				} else {
					return err
				}
			} else if !dir.IsDir() {
				return fmt.Errorf("cannot create directory %q: File exists", fpath)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
				return err
			}

			f, err := os.Create(fpath)
			if err != nil {
				return err
			}
			if _, err := f.Write(file.Body); err != nil {
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}
