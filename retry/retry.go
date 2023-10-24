package retry

import (
	"errors"
	"time"
)

var errNoMoreRetry error = errorNoMoreRetry("no more retry")

type errorNoMoreRetry string

func (err errorNoMoreRetry) Error() string {
	return string(err)
}

func (errorNoMoreRetry) Unwrap() error {
	return errNoMoreRetry
}

// IsNoMoreRetry reports whether error is NoMoreRetry error.
func IsNoMoreRetry(err error) bool {
	if e, ok := err.(interface{ Unwrap() []error }); ok {
		for _, err := range e.Unwrap() {
			if IsNoMoreRetry(err) {
				return true
			}
		}
		return false
	}
	return errors.Is(err, errNoMoreRetry)
}

// ErrNoMoreRetry tells function does no more retry.
func ErrNoMoreRetry(err string) error { return errorNoMoreRetry(err) }

// Do keeps retrying the function until no error is returned.
func Do(fn func() error, attempts, delay int) error {
	var errs []error
	for i := 0; i < attempts; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		errs = append(errs, err)
		if IsNoMoreRetry(err) {
			break
		} else if i < attempts-1 {
			time.Sleep(time.Second * time.Duration(delay))
		}
	}
	return errors.Join(errs...)
}
