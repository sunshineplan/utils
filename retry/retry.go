package retry

import (
	"errors"
	"fmt"
	"time"
)

// ErrNoMoreRetry is a sentinel error indicating that no more retries should be performed.
var ErrNoMoreRetry = errors.New("no more retry")

// StopRetry creates a wrapped error indicating that retries should stop.
func StopRetry(msg string) error {
	return fmt.Errorf("%w: %s", ErrNoMoreRetry, msg)
}

// IsNoMoreRetry reports whether the given error indicates to stop retrying.
func IsNoMoreRetry(err error) bool {
	return errors.Is(err, ErrNoMoreRetry)
}

// Do executes fn repeatedly until it succeeds, the attempts are exhausted,
// or fn returns an error that indicates no more retries.
func Do(fn func() error, attempts int, delay time.Duration) error {
	if attempts <= 0 {
		return errors.New("invalid attempts count")
	}
	var errs []error
	for i := range attempts {
		err := fn()
		if err == nil {
			return nil
		}
		errs = append(errs, err)
		if IsNoMoreRetry(err) {
			break
		}
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	return errors.Join(errs...)
}
