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

	ocrAPI2 = "https://ocr-example.herokuapp.com/file"
)

var errNoResult = errors.New("no ocr result")

// OCR reads image from reader r and converts it into string.
func OCR(r io.Reader) (string, error) {
	return OCRWithClient(r, http.DefaultClient)
}

// OCRWithClient reads image from reader r and converts it into string
// with custom http.Client.
func OCRWithClient(r io.Reader, client *http.Client) (string, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)

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
	w.Close()

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

	var res struct {
		ParsedResults []struct {
			ParsedText string
		}
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return "", err
	}
	if len(res.ParsedResults) == 0 {
		return "", errNoResult
	}
	return res.ParsedResults[0].ParsedText, nil
}

// OCR2 reads image from reader r and converts it into string.
func OCR2(r io.Reader) (string, error) {
	return OCR2WithClient(r, http.DefaultClient)
}

// OCR2WithClient reads image from reader r and converts it into string
// with custom http.Client.
func OCR2WithClient(r io.Reader, client *http.Client) (string, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	part, _ := w.CreateFormFile("file", "pic.jpg")
	if _, err := io.Copy(part, r); err != nil {
		return "", err
	}
	w.Close()

	req, _ := http.NewRequest("POST", ocrAPI2, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var res struct {
		Result  string
		Version string
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return "", err
	}
	if res.Result == "" {
		return "", errNoResult
	}
	return res.Result, nil
}
