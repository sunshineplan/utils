package unit

import (
	"testing"
)

func TestParseByteSize(t *testing.T) {
	if _, err := ParseByteSize("10"); err != nil {
		t.Error(err)
	}
	if _, err := ParseByteSize("10mb"); err != nil {
		t.Error(err)
	}
	if _, err := ParseByteSize("10m"); err != nil {
		t.Error(err)
	}
	if _, err := ParseByteSize("10 MB"); err != nil {
		t.Error(err)
	}
	if _, err := ParseByteSize("10  MB"); err == nil {
		t.Error("expected error; got nil")
	}
	if _, err := ParseByteSize("10MP"); err == nil {
		t.Error("expected error; got nil")
	}
}

func TestByteSize(t *testing.T) {
	for _, testcase := range []struct {
		n   ByteSize
		str string
	}{
		{KB, "1KB"},
		{10 * MB, "10MB"},
		{1536 * MB, "1.5GB"},
		{NewByteSize(1.5, GB), "1.5GB"},
	} {
		if bytesize := ByteSize(testcase.n).String(); bytesize != testcase.str {
			t.Errorf("expected %q; got %q", testcase.str, bytesize)
		}
		if size, err := ParseByteSize(testcase.str); err != nil {
			t.Error(err)
		} else if size != testcase.n {
			t.Errorf("expected %q; got %q", testcase.n, size)
		}
	}
}
