package unit

import (
	"testing"
	"time"
)

func TestFormatBytes(t *testing.T) {
	size := FormatBytes(1024)
	if expect := "1.00 KB"; size != expect {
		t.Errorf("expected %q; got %q", expect, size)
	}
	size = FormatBytes(1024 * 1024)
	if expect := "1.00 MB"; size != expect {
		t.Errorf("expected %q; got %q", expect, size)
	}
	size = FormatBytes(1024 * 1024 * 1024)
	if expect := "1.00 GB"; size != expect {
		t.Errorf("expected %q; got %q", expect, size)
	}
}

func TestFormatDuration(t *testing.T) {
	duration := FormatDuration(30*24*time.Hour + 4*time.Hour + 50*time.Minute + 6*time.Second)
	if expect := "30d4h50m6s"; duration != expect {
		t.Errorf("expected %q; got %q", expect, duration)
	}
	duration = FormatDuration(time.Minute)
	if expect := "1m"; duration != expect {
		t.Errorf("expected %q; got %q", expect, duration)
	}
}
