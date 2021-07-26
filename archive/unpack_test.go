package archive

import (
	"os"
	"reflect"
	"testing"
)

func TestUnPack(t *testing.T) {
	tc := []string{"testdata/test.zip", "testdata/test.tar.gz"}
	result := []File{
		{Name: "1.txt", Body: []byte("1")},
		{Name: "2.txt", Body: []byte("2")},
	}
	for _, i := range tc {
		f, err := os.Open(i)
		if err != nil {
			t.Fatal(err)
		}
		fs, err := Unpack(f)
		if err != nil {
			t.Fatalf("Unpack %q failed: %v", i, err)
		}
		if !reflect.DeepEqual(fs, result) {
			t.Errorf("expected %#v; got %#v", result, fs)
		}
	}

	f, err := os.Open("testdata/folder.zip")
	if err != nil {
		t.Fatal(err)
	}
	if err := UnpackToFiles(f, "testdata"); err == nil {
		t.Error("expected error; got nil")
	}
}
