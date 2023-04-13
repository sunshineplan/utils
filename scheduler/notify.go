package scheduler

import (
	"sync"
	"time"
)

var (
	subscriber = make(map[chan struct{}]struct{})
	mu         sync.Mutex
)

func init() {
	go func() {
		var last time.Time
		for t := range time.NewTicker(time.Second).C {
			if last.IsZero() {
				last = t
			} else {
				if sub := t.Sub(last.Add(time.Second)); sub >= 10*time.Millisecond || sub <= -10*time.Millisecond {
					mu.Lock()
					for k := range subscriber {
						k <- struct{}{}
					}
					mu.Unlock()
				}
				last = t
			}
		}
	}()
}

func subscribeNotify(c chan struct{}) {
	mu.Lock()
	defer mu.Unlock()
	subscriber[c] = struct{}{}
}

func unsubscribeNotify(c chan struct{}) {
	mu.Lock()
	defer mu.Unlock()
	delete(subscriber, c)
}
