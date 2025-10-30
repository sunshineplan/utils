package scheduler

import (
	"time"

	"github.com/sunshineplan/utils/container"
)

// subscriber is a global registry mapping event channels to their target times.
// It is used internally by all running schedulers to receive tick events.
var subscriber = container.NewMap[chan Event, time.Time]()

// Event represents a time event emitted by the scheduler engine.
// It carries both the actual trigger time (Time) and the intended schedule time (Goal).
// If Missed is true, the event was triggered after its scheduled Goal.
type Event struct {
	Time   time.Time
	Goal   time.Time
	Missed bool
}

// init launches a global background goroutine that ticks every second.
// For each tick, it compares the current time with all subscribed times,
// and sends corresponding Event values to their channels.
//
// This mechanism allows multiple Scheduler instances to share the same
// centralized time source and operate concurrently.
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

// subscribe registers a channel to receive an Event when the given time arrives.
func subscribe(t time.Time, c chan Event) {
	subscriber.Swap(c, t.Truncate(time.Second))
}

// unsubscribe removes a previously registered channel from the subscriber map.
func unsubscribe(c chan Event) {
	subscriber.Delete(c)
}
