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

func TestNewByteSize(t *testing.T) {
	tests := []struct {
		n         float64
		unit      string
		want      ByteSize
		expectErr bool
	}{
		{1, "B", B, false},
		{1, "KB", KB, false},
		{1, "mb", MB, false},
		{1.5, "GB", ByteSize(1.5 * float64(GB)), false},
		{2, "tb", ByteSize(2 * float64(TB)), false},
		{0, "MB", 0, false},
		{1, "XYZ", 0, true},
	}

	for _, tt := range tests {
		got, err := NewByteSize(tt.n, tt.unit)
		if tt.expectErr {
			if err == nil {
				t.Errorf("NewByteSize(%v, %q) expected error, got nil", tt.n, tt.unit)
			}
			continue
		}
		if err != nil {
			t.Errorf("NewByteSize(%v, %q) unexpected error: %v", tt.n, tt.unit, err)
			continue
		}
		if got != tt.want {
			t.Errorf("NewByteSize(%v, %q) = %v, want %v", tt.n, tt.unit, got, tt.want)
		}
	}
}

func TestTo(t *testing.T) {
	tests := []struct {
		size      ByteSize
		unit      string
		decimals  int
		want      string
		expectErr bool
	}{
		// integer results
		{ByteSize(1024), "KB", 0, "1KB", false},
		{ByteSize(1536), "KB", 0, "2KB", false},
		{ByteSize(1048576), "MB", 0, "1MB", false},

		// fractional results
		{ByteSize(1536), "KB", 2, "1.5KB", false},
		{ByteSize(1536), "MB", 3, "0.001MB", false},
		{MustParseByteSize("1.5GB"), "GB", 2, "1.5GB", false},

		// small numbers < 1
		{ByteSize(1), "KB", 6, "0.000977KB", false},
		{ByteSize(500), "MB", 6, "0.000477MB", false},

		// decimals = 0 for fractional -> rounds to nearest int
		{ByteSize(1536), "KB", 0, "2KB", false},

		// invalid unit
		{ByteSize(1024), "XYZ", 2, "", true},
	}

	for _, tt := range tests {
		got, err := tt.size.To(tt.unit, tt.decimals)
		if tt.expectErr {
			if err == nil {
				t.Errorf("To(%q, %d) expected error, got nil", tt.unit, tt.decimals)
			}
			continue
		}
		if err != nil {
			t.Errorf("To(%q, %d) unexpected error: %v", tt.unit, tt.decimals, err)
			continue
		}
		if got != tt.want {
			t.Errorf("To(%q, %d) = %q, want %q", tt.unit, tt.decimals, got, tt.want)
		}
	}
}
