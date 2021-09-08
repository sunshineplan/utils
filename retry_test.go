package utils

import (
	"errors"
	"testing"
)

func TestRetry(t *testing.T) {
	if err := Retry(func() error {
		return nil
	}, 3, 1); err != nil {
		t.Errorf("expected nil error; got non-nil error %v", err)
	}

	if err := Retry(func() error {
		return errors.New("error")
	}, 3, 1); err == nil {
		t.Error("expected non-nil error; got nil error")
	}
}
