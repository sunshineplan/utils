package utils

import (
	"fmt"
	"io"
	"net/http"

	"github.com/sunshineplan/utils/executor"
)

// UserAgentString gets latest chrome user agent string.
func UserAgentString() (string, error) {
	result, err := executor.ExecuteConcurrentArg(
		[]string{
			"https://raw.githubusercontent.com/sunshineplan/useragent/main/README.md",
			"https://cdn.jsdelivr.net/gh/sunshineplan/useragent/README.md",
			"https://fastly.jsdelivr.net/gh/sunshineplan/useragent/README.md",
		},
		func(url string) (any, error) {
			return http.Get(url)
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to fetch user agent string: %s", err)
	}

	resp := result.(*http.Response)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("no StatusOK response")
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// UserAgent gets latest chrome user agent string, if failed to get string or
// string is empty, the default string will be used.
func UserAgent(defaultUserAgentString string) string {
	ua, err := UserAgentString()
	if err != nil || ua == "" {
		ua = defaultUserAgentString
	}

	return ua
}
