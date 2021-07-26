package utils

import (
	"errors"
	"time"
)

// ErrNoMoreRetry tells function does no more retry.
var ErrNoMoreRetry = errors.New("no more retry")

// Retry keeps retrying the function until no error is returned.
func Retry(fn func() error, attempts, delay uint) (err error) {
	for i := uint(0); i < attempts; i++ {
		if err = fn(); err == nil || err == ErrNoMoreRetry {
			return
		}

		if i < attempts-1 {
			time.Sleep(time.Second * time.Duration(delay))
		}
	}

	return
}
