package ocr

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func TestOCR(t *testing.T) {
	apiKey, ok := os.LookupEnv("OCR_APIKEY")
	if !ok || apiKey == "" {
		t.Skip("OCR_APIKEY is not set; skipping OCR integration test")
	}

	prevKey := APIKey
	APIKey = apiKey
	defer func() {
		APIKey = prevKey
	}()

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
