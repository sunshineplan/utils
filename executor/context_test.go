package executor

import (
	"errors"
	"testing"
)

func TestSkip(t *testing.T) {
	tmp := errors.New("error")
	if _, err := ExecuteSerial(
		[]int{0, 1, 2},
		func(n int) (any, error) {
			if n == 0 {
				return nil, tmp
			}
			return nil, SkipErr
		},
	); err != tmp {
		t.Errorf("expected %d; got %v", tmp, err)
	}

	if _, err := ExecuteSerial(
		[]int{0, 1, 2},
		func(n int) (any, error) {
			return nil, SkipErr
		},
	); err != AllSkippedErr {
		t.Errorf("expected %s; got %v", AllSkippedErr, err)
	}
}
