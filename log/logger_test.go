package log

import (
	"bytes"
	"log"
	"log/slog"
	"os"
	"runtime"
	"testing"
)

func TestLogger(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	logger := New("test1", "", 0)
	defer os.Remove("test1")
	if file := logger.File(); file != "test1" {
		t.Errorf("expected test1; got %q", file)
	}
	logger.Print("test1")
	if err := os.Rename("test1", "test2"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test2")
	logger.Rotate()
	if file := logger.File(); file != "test1" {
		t.Errorf("expected test1; got %q", file)
	}
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
		t.Errorf("expected test2; got %q", s)
	}
	if s := string(b2); s != "test1\n" {
		t.Errorf("expected test1; got %q", s)
	}
}

func TestSLogger(t *testing.T) {
	var buf bytes.Buffer
	l := newLogger(log.New(&buf, "", 0), nil)
	if file := l.File(); file != "" {
		t.Errorf("expected empty string; got %q", file)
	}
	l.Info("test")
	if s, expected := buf.String(), "INFO test\n"; s != expected {
		t.Errorf("expected %q; got %q", expected, s)
	}
	buf.Reset()
	l.Debug("test")
	if s, expected := buf.String(), ""; s != expected {
		t.Errorf("expected %q; got %q", expected, s)
	}
	buf.Reset()
	l.SetLevel(slog.LevelDebug)
	l.Debug("test")
	if s, expected := buf.String(), "DEBUG test\n"; s != expected {
		t.Errorf("expected %q; got %q", expected, s)
	}
	buf.Reset()
	l = l.WithGroup("g").With("a", 1)
	l.Info("test")
	if s, expected := buf.String(), "INFO test g.a=1\n"; s != expected {
		t.Errorf("expected %q; got %q", expected, s)
	}
}
