package utils

import (
	"errors"
	"strconv"
	"testing"
)

func TestRetry(t *testing.T) {
	var i int
	if err := Do(func() error {
		defer func() { i++ }()
		return nil
	}, 3, 1); err != nil {
		t.Errorf("expected nil error; got non-nil error %v", err)
	} else if i != 1 {
		t.Errorf("expected 1; got %d", i)
	}

	i = 0
	if err := Do(func() error {
		defer func() { i++ }()
		return errors.New("error" + strconv.Itoa(i))
	}, 3, 1); err == nil {
		t.Error("expected non-nil error; got nil error")
	} else if err.Error() != "error2" {
		t.Errorf("expected error2; got %s", err)
	}

	i = 0
	if err := Do(func() error {
		defer func() { i++ }()
		return ErrNoMoreRetry("error" + strconv.Itoa(i))
	}, 3, 1); !IsNoMoreRetry(err) {
		t.Errorf("expected ErrNoMoreRetry; got %s", err)
	} else if err.Error() != "error0" {
		t.Errorf("expected error0; got %d", i)
	}
}
