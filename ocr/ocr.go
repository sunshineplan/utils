package ocr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

const (
	ocrAPI    = "https://api.ocr.space/parse/image"
	ocrAPIKey = "5a64d478-9c89-43d8-88e3-c65de9999580"
)

var errNoResult = errors.New("no ocr result")

type ocrResponse struct {
	ParsedResults []struct {
		ParsedText string
	}
}

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

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ocr request failed: %s", resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var res ocrResponse
	if err := json.Unmarshal(b, &res); err != nil {
		return "", err
	}
	if len(res.ParsedResults) == 0 {
		return "", errNoResult
	}
	return res.ParsedResults[0].ParsedText, nil
}
