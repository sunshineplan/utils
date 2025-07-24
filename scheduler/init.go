package scheduler

import (
	"time"

	"github.com/sunshineplan/utils/container"
)

var subscriber = container.NewMap[chan time.Time, time.Time]()

func init() {
	go func() {
		for t := range time.NewTicker(time.Second).C {
			go subscriber.Range(func(k chan time.Time, v time.Time) bool {
				if v.Equal(t.Truncate(time.Second)) {
					k <- v
				}
				return true
			})
		}
	}()
}

func subscribe(t time.Time, c chan time.Time) {
	subscriber.Store(c, t.Truncate(time.Second))
}

func unsubscribe(c chan time.Time) {
	subscriber.Delete(c)
}
