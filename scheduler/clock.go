package scheduler

import (
	"fmt"
	"time"
)

type Clock struct {
	Hr, Min, Sec int
}

func AtClock(hour, min, sec int) *Clock {
	if hour > 23 || hour < 0 ||
		min > 59 || min < 0 ||
		sec > 59 || sec < 0 {
		panic(fmt.Sprintf("invalid clock: hour(%d) min(%d) sec(%d)", hour, min, sec))
	}
	return &Clock{hour, min, sec}
}

func Hour(hour int) *Clock {
	return AtClock(hour, 0, 0)
}

func Minute(min int) *Clock {
	return AtClock(0, min, 0)
}

func Second(sec int) *Clock {
	return AtClock(0, 0, sec)
}

func (c *Clock) Hour(hour int) *Clock {
	if hour > 23 || hour < 0 {
		panic(fmt.Sprintln("invalid hour", hour))
	}
	c.Hr = hour
	return c
}

func (c *Clock) Minute(min int) *Clock {
	if min > 59 || min < 0 {
		panic(fmt.Sprintln("invalid minute", min))
	}
	c.Min = min
	return c
}

func (c *Clock) Second(sec int) *Clock {
	if sec > 59 || sec < 0 {
		panic(fmt.Sprintln("invalid second", sec))
	}
	c.Sec = sec
	return c
}

func (c Clock) IsMatched(t time.Time) bool {
	hour, min, sec := t.Clock()
	return c.Hr == hour && c.Min == min && c.Sec == sec
}
