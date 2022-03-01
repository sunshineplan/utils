package utils

import (
	"time"
)

var _ error = errNoMoreRetry("")

type errNoMoreRetry string

func (err errNoMoreRetry) Error() string {
	return string(err)
}

// IsNoMoreRetry reports whether error is NoMoreRetry error.
func IsNoMoreRetry(err error) bool {
	_, ok := err.(errNoMoreRetry)
	return ok
}

// ErrNoMoreRetry tells function does no more retry.
func ErrNoMoreRetry(err string) error { return errNoMoreRetry(err) }

// Do keeps retrying the function until no error is returned.
func Do(fn func() error, attempts, delay int) (err error) {
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil || IsNoMoreRetry(err) {
			return
		}

		if i < attempts-1 {
			time.Sleep(time.Second * time.Duration(delay))
		}
	}

	return
}
