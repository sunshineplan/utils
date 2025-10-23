package progressbar

import (
	"fmt"
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

	pb := New(15).SetRefreshInterval(4 * time.Second)
	pb.Start()
	pb.Additional("refreshes in 4s")
	for range pb.total {
		pb.Add(1)
		time.Sleep(time.Second)
	}
	pb.Wait()

	pb = New(10).SetRefreshInterval(500 * time.Millisecond)
	pb.Start()
	pb.Additional("refreshes in 500ms")
	for range pb.total {
		pb.Add(1)
		time.Sleep(time.Second)
	}
	pb.Wait()

	pb = New(0)
	pb.Start()
}

func TestMessage(t *testing.T) {
	pb := New(15)
	pb.Start()
	errCh := make(chan error, 1)
	stopCh := make(chan struct{})
	go func() {
		i := 0
		for {
			select {
			case <-stopCh:
				return
			default:
				time.Sleep(500 * time.Millisecond)
				if err := pb.Message(fmt.Sprintf("test messages (%d)", i)); err != nil {
					errCh <- err
					return
				}
			}
			i++
		}
	}()
	for range pb.total {
		pb.Add(1)
		time.Sleep(time.Second)
	}
	pb.Wait()
	close(stopCh)
	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
	if err := pb.Message("test messages"); err == nil {
		t.Fatal("expected non-nil error; got nil error")
	}
}

func TestCancel(t *testing.T) {
	pb := New(15).SetRefreshInterval(4 * time.Second)
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
	pb.Wait()
}

func TestFromReader(t *testing.T) {
	resp, err := http.Get("https://github.com/sunshineplan/feh/releases/latest/download/feh")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	total, err := strconv.ParseInt(resp.Header.Get("content-length"), 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	pb := New(total).SetUnit("bytes")
	if _, err := pb.FromReader(resp.Body, io.Discard); err != nil {
		t.Fatal(err)
	}
	pb.Wait()
}

func TestSetTemplate(t *testing.T) {
	pb := &ProgressBar[int]{}
	if err := pb.SetTemplate(`{{.Done}}`); err != nil {
		t.Error(err)
	}
	if err := pb.SetTemplate(`{{.Test}}`); err == nil {
		t.Error("expected non-nil error; got nil error")
	}
}
