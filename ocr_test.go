package utils

import (
	"log"
	"os"
	"testing"
)

func TestOCR(t *testing.T) {
	f, err := os.Open("testdata/ocr.space.logo.png")
	if err != nil {
		t.Fatal(err)
	}
	r, err := OCR(f)
	if err != nil {
		if err.Error() == "no ocr result" {
			log.Print("maybe ocr api is down")
			return
		}
		t.Fatal(err)
	}
	if expect := "OCR .Space\r\n"; r != expect {
		t.Fatalf("expected %q; got %q", expect, r)
	}
}
