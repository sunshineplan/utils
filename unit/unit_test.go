package unit

import (
	"testing"
	"time"
)

func TestFormatBytes(t *testing.T) {
	for _, testcase := range []struct {
		n      uint64
		expect string
	}{
		{1024, "1.00 KB"},
		{1024 * 1024, "1.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
	} {
		if size := FormatBytes(testcase.n); size != testcase.expect {
			t.Errorf("expected %q; got %q", testcase.expect, size)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	for _, testcase := range []struct {
		d      time.Duration
		expect string
	}{
		{30*24*time.Hour + 4*time.Hour + 50*time.Minute + 6*time.Second, "30d4h50m6s"},
		{time.Minute, "1m"},
		{0, "0s"},
	} {
		if duration := FormatDuration(testcase.d); duration != testcase.expect {
			t.Errorf("expected %q; got %q", testcase.expect, duration)
		}
	}
}
