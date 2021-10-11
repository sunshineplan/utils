package watcher

import (
	"log"
	"os"
	"testing"
	"time"
)

func TestWatcher(t *testing.T) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	f.WriteString("test")
	f.Close()

	w := New(f.Name(), time.Second)
	if w.File() != f.Name() {
		t.Fatal("filename is not same")
	}

	f, err = os.Create(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("changed")
	f.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	select {
	case <-w.C:
		log.Print("file changed")
	case <-ticker.C:
		t.Fatal("timeout")
	}

	f, err = os.Create(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("test")
	f.Close()

	ticker.Reset(2 * time.Second)

	select {
	case <-w.C:
		log.Print("file changed")
	case <-ticker.C:
		t.Fatal("timeout")
	}

	w.Stop()

	f, err = os.Create(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("stoped")
	f.Close()

	ticker.Reset(2 * time.Second)

	select {
	case <-w.C:
		t.Fatal("watcher failed to stop")
	case <-ticker.C:
		log.Print("stoped")
	}
}
