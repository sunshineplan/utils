package utils

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func TestOCR(t *testing.T) {
	b, err := os.ReadFile("testdata/ocr.space.logo.png")
	if err != nil {
		t.Fatal(err)
	}

	r1, err := OCR(bytes.NewReader(b))
	if err != nil {
		if err == errNoResult {
			log.Print("maybe ocr api is down")
			return
		}
		t.Fatal(err)
	}
	if expect := "OCR Space\r\n"; r1 != expect {
		t.Errorf("expected %q; got %q", expect, r1)
	}
}
