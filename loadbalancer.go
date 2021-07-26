package utils

import (
	"errors"
	"sync"
)

// LoadBalancer gets the fastest result from the same function use several selector
func LoadBalancer(selector []interface{}, fn func(interface{}, chan<- interface{}, chan<- error)) (interface{}, error) {
	count := len(selector)
	if count == 0 {
		return nil, errors.New("selector can't be empty")
	}

	var mu sync.Mutex
	result := make(chan interface{}, 1)
	lasterr := make(chan error, 1)
	done := make(chan bool, 1)

	run := func(s interface{}) {
		rc := make(chan interface{}, 1)
		ec := make(chan error, 1)

		go fn(s, rc, ec)

		for {
			select {
			case ok := <-done:
				if ok {
					return
				}
			case err := <-ec:
				mu.Lock()

				if count == 1 {
					result <- nil
					lasterr <- err
					done <- false
				}
				count--

				mu.Unlock()

				return
			case r := <-rc:
				result <- r
				done <- true

				return
			}
		}
	}

	for _, i := range selector {
		go run(i)
	}

	r := <-result
	if len(lasterr) == 0 {
		return r, nil
	}

	return nil, <-lasterr
}
