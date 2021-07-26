package utils

import (
	"net/http"
	"os"
	"testing"
)

func TestOCR(t *testing.T) {
	f, err := os.Open("testdata/ocr.space.logo.png")
	if err != nil {
		t.Error(err)
	}
	r, err := OCRWithClient(f, http.DefaultClient)
	if err != nil {
		t.Error(err)
	}
	if r != "OCR .Space\r\n" {
		t.Errorf("expected %q; got %q", "OCR .Space\r\n", r)
	}
}
