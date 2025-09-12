package archive

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

// Format represents the archive format.
type Format int

const (
	// ZIP format
	ZIP Format = iota
	// TAR format
	TAR
)

type format struct {
	format Format
	magic  string
}

var formats = []format{
	{ZIP, zipMagic},
	{TAR, tarMagic},
}

// ErrFormat indicates that encountered an unknown format.
var ErrFormat = errors.New("unknown format")

type reader interface {
	io.Reader
	Peek(int) ([]byte, error)
}

func asReader(r io.Reader) reader {
	if rr, ok := r.(reader); ok {
		return rr
	}
	return bufio.NewReader(r)
}

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

func isArchive(r reader) (bool, Format) {
	for _, f := range formats {
		b, err := r.Peek(len(f.magic))
		if err == nil && match(f.magic, b) {
			return true, f.format
		}
	}
	return false, -1
}

// IsArchive tests b is an archive file or not, if ok also return its format.
func IsArchive(b []byte) (bool, Format) {
	return isArchive(asReader(bytes.NewReader(b)))
}

// File struct contains bytes body and the provided name field.
type File struct {
	Name  string
	Body  []byte
	IsDir bool
}
