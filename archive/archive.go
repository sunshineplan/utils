package archive

import (
	"errors"
	"os"
)

// Format represents the archive format.
type Format int

const (
	// ZIP format
	ZIP Format = iota
	// TAR format
	TAR
)

// ErrFormat indicates that encountered an unknown format.
var ErrFormat = errors.New("unknown format")

func match(magic string, b []byte) bool {
	if len(magic) != len(b) {
		return false
	}
	for i, c := range b {
		if magic[i] != c && magic[i] != '?' {
			return false
		}
	}
	return true
}

// IsArchive tests b is an archive file or not, if ok also return its format.
func IsArchive(b []byte) (bool, Format) {
	switch {
	case match(zipMagic, b[:len(zipMagic)]):
		return true, ZIP
	case match(tarMagic, b[:len(tarMagic)]):
		return true, TAR
	default:
		return false, -1
	}
}

// File struct contains bytes body and the provided name field.
type File struct {
	Name  string
	Body  []byte
	IsDir bool
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
