package log

import (
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	logger := New("test1", "", 0)
	defer os.Remove("test1")
	logger.Print("test1")
	if err := os.Rename("test1", "test2"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test2")
	logger.Rotate()
	logger.Print("test2")
	b1, err := os.ReadFile("test1")
	if err != nil {
		t.Fatal(err)
	}
	b2, err := os.ReadFile("test2")
	if err != nil {
		t.Fatal(err)
	}
	if s := string(b1); s != "test2\n" {
		t.Errorf("expected test2; got %s", s)
	}
	if s := string(b2); s != "test1\n" {
		t.Errorf("expected test1; got %s", s)
	}
}
