package utils

import (
	"fmt"
	"io"
	"net/http"

	"github.com/sunshineplan/utils/executor"
)

const uaAPI = "https://raw.githubusercontent.com/sunshineplan/useragent/main/user-agent"
const uaCDNAPI = "https://cdn.jsdelivr.net/gh/sunshineplan/useragent/user-agent"

// UserAgentString gets latest chrome user agent string.
func UserAgentString() (string, error) {
	result, err := executor.ExecuteConcurrentArg(
		[]string{uaAPI, uaCDNAPI},
		func(url interface{}) (interface{}, error) {
			return http.Get(url.(string))
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
