package scheduler

import (
	"time"

	"github.com/sunshineplan/utils/container"
)

var subscriber = container.NewMap[chan Event, time.Time]()

type Event struct {
	Time   time.Time
	Goal   time.Time
	Missed bool
}

func init() {
	go func() {
		for t := range time.NewTicker(time.Second).C {
			t = t.Truncate(time.Second)
			go subscriber.Range(func(k chan Event, v time.Time) bool {
				switch v.Compare(t) {
				case -1:
					k <- Event{t, v, true}
				case 0:
					k <- Event{t, v, false}
				}
				return true
			})
		}
	}()
}

func subscribe(t time.Time, c chan Event) {
	subscriber.Swap(c, t.Truncate(time.Second))
}

func unsubscribe(c chan Event) {
	subscriber.Delete(c)
}
