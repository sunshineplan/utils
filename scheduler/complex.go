package scheduler

import (
	"fmt"
	"strings"
	"time"
)

var (
	_ Schedule = complexSched(nil)

	_ complexSched = multiSched{}
	_ complexSched = condSched{}
)

// complexSched defines an internal interface for composite schedules.
// It extends Schedule with initialization and introspection capabilities.
type complexSched interface {
	// IsMatched reports whether the given time matches the composite condition.
	IsMatched(time.Time) bool
	// Next returns the next time that satisfies the composite condition.
	Next(time.Time) time.Time
	// String returns a human-readable representation.
	String() string

	// init initializes internal states or nested schedules using the given start time.
	init(t time.Time)
	// len returns the number of sub-schedules contained in this composite schedule.
	len() int
}

// complex is a type constraint used for generic initialization of schedule slices.
type complex interface {
	~[]Schedule
}

// initComplexSched initializes all sub-schedules that implement complexSched or tickerSched.
// It ensures that nested or periodic schedules have their starting point properly set.
func initComplexSched[sche complex](s sche, t time.Time) {
	for _, s := range s {
		if i, ok := s.(complexSched); ok {
			i.init(t)
		} else if i, ok := s.(*tickerSched); ok {
			i.init(t)
		}
	}
}

// multiSched represents a composite schedule that matches if *any*
// of its sub-schedules match — i.e., a logical OR operation.
type multiSched []Schedule

// MultiSchedule creates a new schedule that triggers when any of the provided
// schedules match. Equivalent to a logical OR of all schedules.
func MultiSchedule(schedules ...Schedule) Schedule {
	return multiSched(schedules)
}

// init initializes nested schedules recursively.
func (s multiSched) init(t time.Time) {
	initComplexSched(s, t)
}

// len returns the number of contained sub-schedules.
func (s multiSched) len() int {
	return len(s)
}

// IsMatched returns true if any sub-schedule matches the given time.
func (s multiSched) IsMatched(t time.Time) bool {
	for _, i := range s {
		if i.IsMatched(t) {
			return true
		}
	}
	return false
}

// Next returns the earliest next time among all sub-schedules.
// If no valid next time exists, it returns a zero time.
func (s multiSched) Next(t time.Time) (next time.Time) {
	for _, i := range s {
		if t := i.Next(t); next.IsZero() || !t.IsZero() && t.Before(next) {
			next = t
		}
	}
	return
}

// String returns a readable representation of the multi-schedule.
func (s multiSched) String() string {
	switch len(s) {
	case 0:
		return ""
	case 1:
		return s[0].String()
	default:
		var b strings.Builder
		b.WriteString("MultiSchedule: ")
		for i, sched := range s {
			if i != 0 {
				fmt.Fprint(&b, "; ")
			}
			fmt.Fprint(&b, sched)
		}
		return b.String()
	}
}

// condSched represents a composite schedule that matches only if *all*
// of its sub-schedules match — i.e., a logical AND operation.
type condSched []Schedule

// ConditionSchedule creates a new schedule that triggers only when all
// of the provided schedules match simultaneously. Equivalent to a logical AND.
func ConditionSchedule(schedules ...Schedule) Schedule {
	return condSched(schedules)
}

// init initializes nested schedules recursively.
func (s condSched) init(t time.Time) {
	initComplexSched(s, t)
}

// len returns the number of contained sub-schedules.
func (s condSched) len() int {
	return len(s)
}

// IsMatched returns true only if all sub-schedules match the given time.
func (s condSched) IsMatched(t time.Time) bool {
	if s.len() == 0 {
		return false
	}
	for _, i := range s {
		if !i.IsMatched(t) {
			return false
		}
	}
	return true
}

// Next returns the next time that satisfies all sub-schedules simultaneously.
// If there are no schedules, it returns zero time.
// If the current time already matches, it advances by one second to find the next occurrence.
func (s condSched) Next(t time.Time) (next time.Time) {
	if l := len(s); l == 0 {
		return time.Time{}
	} else if l == 1 {
		return s[0].Next(t)
	}
	// Avoid returning the same time repeatedly if it already matches.
	if s.IsMatched(t) {
		t = t.Add(time.Second)
	}
	// Increment one second at a time until all conditions match,
	// but limit the search to one year to avoid infinite loops.
	for next = t.Truncate(time.Second); !s.IsMatched(next); next = next.Add(time.Second) {
		if next.Sub(t) >= time.Hour*24*366 {
			return time.Time{}
		}
	}
	return
}

// String returns a readable representation of the condition schedule.
func (s condSched) String() string {
	switch len(s) {
	case 0:
		return ""
	case 1:
		return s[0].String()
	default:
		var b strings.Builder
		b.WriteString("ConditionSchedule: ")
		for i, sched := range s {
			if i != 0 {
				fmt.Fprint(&b, "; ")
			}
			fmt.Fprint(&b, sched)
		}
		return b.String()
	}
}
