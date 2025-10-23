package retry

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	if err := StopRetry("error"); !errors.Is(err, ErrNoMoreRetry) {
		t.Error("expected err is ErrNoMoreRetry; got not")
	}
	var i int
	if err := Do(func() error {
		defer func() { i++ }()
		return nil
	}, 3, time.Second); err != nil {
		t.Errorf("expected nil error; got non-nil error %v", err)
	} else if i != 1 {
		t.Errorf("expected 1; got %d", i)
	}

	i = 0
	if err := Do(func() error {
		defer func() { i++ }()
		return errors.New("error" + strconv.Itoa(i))
	}, 3, time.Second); err == nil {
		t.Error("expected non-nil error; got nil error")
	} else if expect := "error0\nerror1\nerror2"; err.Error() != expect {
		t.Errorf("expected %s; got %s", expect, err)
	}

	i = 0
	if err := Do(func() error {
		defer func() { i++ }()
		return StopRetry("error" + strconv.Itoa(i))
	}, 3, time.Second); !IsNoMoreRetry(err) {
		t.Errorf("expected ErrNoMoreRetry; got %s", err)
	} else if err.Error() != "no more retry: error0" {
		t.Errorf("expected error0; got %s", err)
	}
}
