package archive

import "os"

// File struct contains bytes body and the provided name field.
type File struct {
	Name  string
	Body  []byte
	IsDir bool
}

// Format represents the archive format.
type Format int

const (
	// ZIP format
	ZIP Format = iota
	// TAR format
	TAR
)

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
