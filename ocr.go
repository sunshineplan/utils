package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
)

const (
	ocrAPI    = "https://api.ocr.space/parse/image"
	ocrAPIKey = "5a64d478-9c89-43d8-88e3-c65de9999580"
)

// OCR reads image from reader r and converts it into string.
func OCR(r io.Reader) (string, error) {
	return OCRWithClient(r, &http.Client{Transport: &http.Transport{Proxy: nil}})
}

// OCRWithClient reads image from reader r and converts it into string
// with custom http.Client.
func OCRWithClient(r io.Reader, client *http.Client) (string, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer w.Close()

	part, _ := w.CreateFormFile("file", "pic.jpg")
	if _, err := io.Copy(part, r); err != nil {
		return "", err
	}

	params := map[string]string{
		"scale": "true",
	}
	for k, v := range params {
		w.WriteField(k, v)
	}

	req, _ := http.NewRequest("POST", ocrAPI, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("apikey", ocrAPIKey)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var results struct {
		ParsedResults []struct {
			ParsedText string
		}
	}
	if err := json.Unmarshal(b, &results); err != nil {
		return "", err
	}
	if len(results.ParsedResults) == 0 {
		return "", errors.New("no ocr result")
	}
	return results.ParsedResults[0].ParsedText, nil
}
