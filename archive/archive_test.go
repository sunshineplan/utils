package archive

import "testing"

func TestIsArchive(t *testing.T) {
	if ok, _ := IsArchive([]byte{}); ok {
		t.Error("expected not ok; got ok")
	}
	if ok, _ := IsArchive([]byte("\x1f\x8b\x08")); ok {
		t.Error("expected not ok; got ok")
	}
	if _, format := IsArchive([]byte("\x1f\x8b\x08\x00")); format != TAR {
		t.Errorf("expected format is TAR(%d); got %d", TAR, format)
	}
}
