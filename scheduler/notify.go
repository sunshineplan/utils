package scheduler

import (
	"sync"
	"time"
)

var (
	subscriber = make(map[chan time.Time]struct{})
	mu         sync.Mutex
)

func init() {
	go func() {
		var last time.Time
		for t := range time.NewTicker(time.Second).C {
			if last.IsZero() {
				last = t
			} else {
				sub := t.Sub(last)
				last = t
				if sub := int64(sub / time.Millisecond); sub != 1000 && sub != 999 {
					mu.Lock()
					for k := range subscriber {
						k <- t
					}
					mu.Unlock()
				}
			}
		}
	}()
}

func subscribeNotify(c chan time.Time) {
	mu.Lock()
	defer mu.Unlock()
	subscriber[c] = struct{}{}
}

func unsubscribeNotify(c chan time.Time) {
	mu.Lock()
	defer mu.Unlock()
	delete(subscriber, c)
}
