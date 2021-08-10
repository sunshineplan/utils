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
		t.Error("expected equal bytes; got not equal")
	}
	if _, format := IsArchive(buf1.Bytes()); format != ZIP {
		t.Errorf("expected format is ZIP(%d); got %d", ZIP, format)
	}
}

func TestPackTAR(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	if err := Pack(&buf1, TAR, files...); err != nil {
		t.Fatal(err)
	}
	if _, format := IsArchive(buf1.Bytes()); format != TAR {
		t.Errorf("expected format is TAR(%d); got %d", TAR, format)
	}
	if err := PackFromFiles(&buf2, TAR, "testdata/1.txt", "testdata/2.txt"); err != nil {
		t.Fatal(err)
	}
	if _, format := IsArchive(buf2.Bytes()); format != TAR {
		t.Errorf("expected format is TAR(%d); got %d", TAR, format)
	}
}
