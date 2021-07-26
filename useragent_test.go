package utils

import (
	"strings"
	"testing"
)

func TestUserAgentString(t *testing.T) {
	ua, err := UserAgentString()
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(ua, "Chrome") {
		t.Errorf("expected contains %q; got %q", "Chrome", ua)
	}
}
