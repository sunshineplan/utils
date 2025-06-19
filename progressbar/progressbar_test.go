package progressbar

import (
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestProgessBar(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("gave no panic; want panic")
		}
	}()

	pb := New(15).SetRefresh(4 * time.Second)
	pb.Start()
	pb.Additional("refreshes in 4s")
	for range pb.total {
		pb.Add(1)
		time.Sleep(time.Second)
	}
	pb.Done()

	pb = New(10).SetRefresh(500 * time.Millisecond)
	pb.Start()
	pb.Additional("refreshes in 500ms")
	for range pb.total {
		pb.Add(1)
		time.Sleep(time.Second)
	}
	pb.Done()

	pb = New(0)
	pb.Start()
}

func TestCancel(t *testing.T) {
	pb := New(15).SetRefresh(4 * time.Second)
	pb.Start()
	go func() {
		time.Sleep(3 * time.Second)
		pb.Cancel()
	}()
	go func() {
		for range pb.total {
			pb.Add(1)
			time.Sleep(time.Second)
		}
	}()
	pb.Done()
}

func TestFromReader(t *testing.T) {
	resp, err := http.Get("https://github.com/sunshineplan/feh/releases/latest/download/feh")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	total, err := strconv.Atoi(resp.Header.Get("content-length"))
	if err != nil {
		t.Fatal(err)
	}
	pb := New(total).SetUnit("bytes")
	if _, err := pb.FromReader(resp.Body, io.Discard); err != nil {
		t.Fatal(err)
	}
	pb.Done()
}

func TestSetTemplate(t *testing.T) {
	pb := &ProgressBar{}
	if err := pb.SetTemplate(`{{.Done}}`); err != nil {
		t.Error(err)
	}
	if err := pb.SetTemplate(`{{.Test}}`); err == nil {
		t.Error("expected non-nil error; got nil error")
	}
}
