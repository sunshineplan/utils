package archive

import (
	"bytes"
	"testing"
)

var files = []File{
	{Name: "testdata/1.txt", Body: []byte("1")},
	{Name: "testdata/2.txt", Body: []byte("2")},
}

func TestPackZIP(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	if err := Pack(&buf1, ZIP, files...); err != nil {
		t.Fatal(err)
	}
	if err := PackFromFiles(&buf2, ZIP, "testdata/1.txt", "testdata/2.txt"); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
		t.Error("expected equal zip archive; got not equal")
	}
	if !match(zipMagic, buf1.Bytes()[:len(zipMagic)]) {
		t.Error("expected equal magic; got not equal")
	}
}

func TestPackTAR(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	if err := Pack(&buf1, TAR, files...); err != nil {
		t.Fatal(err)
	}
	if !match(tarMagic, buf1.Bytes()[:len(tarMagic)]) {
		t.Error("expected equal magic; got not equal")
	}
	if err := PackFromFiles(&buf2, TAR, "testdata/1.txt", "testdata/2.txt"); err != nil {
		t.Fatal(err)
	}
	if !match(tarMagic, buf2.Bytes()[:len(tarMagic)]) {
		t.Error("expected equal magic; got not equal")
	}
}
