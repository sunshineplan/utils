package utils

import (
	"context"
	"errors"
	"sync"
)

// LoadBalancer gets the fastest result from the same function use several selector
func LoadBalancer(selector []interface{}, fn func(interface{}) (interface{}, error)) (interface{}, error) {
	count := len(selector)
	if count == 0 {
		return nil, errors.New("selector can't be empty")
	}

	var mu sync.Mutex
	result := make(chan interface{}, 1)
	lasterr := make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	run := func(s interface{}) {
		var r interface{}
		var err error
		c := make(chan error, 1)

		go func() {
			r, err = fn(s)
			c <- err
		}()

		select {
		case <-ctx.Done():
			return
		case err := <-c:
			if err != nil {
				mu.Lock()

				if count == 1 {
					result <- nil
					lasterr <- err
				}
				count--

				mu.Unlock()
			} else {
				cancel()
				result <- r
				lasterr <- nil
			}
		}
	}

	for _, i := range selector {
		go run(i)
	}

	if err := <-lasterr; err != nil {
		return nil, err
	}

	return <-result, nil
}
