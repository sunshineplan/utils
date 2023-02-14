package scheduler

import (
	"fmt"
	"time"
)

var (
	_ Schedule = Clock{}
	_ Schedule = clockSched{}
)

var clockLayout = []string{
	"15:04",
	"15:04:05",
}

type Clock struct {
	hour, min, sec int
}

func AtClock(hour, min, sec int) *Clock {
	if hour > 23 || hour < -1 ||
		min > 59 || min < -1 ||
		sec > 59 || sec < -1 {
		panic(fmt.Sprintf("invalid clock: hour(%d) min(%d) sec(%d)", hour, min, sec))
	}
	return &Clock{hour, min, sec}
}

var FullClock = AtClock(-1, -1, -1)

func AtHour(hour int) *Clock {
	return AtClock(hour, 0, 0)
}

func AtMinute(min int) *Clock {
	return AtClock(-1, min, 0)
}

func AtSecond(sec int) *Clock {
	return AtClock(-1, -1, sec)
}

func ClockFromString(str string) *Clock {
	t, err := parseTime(str, clockLayout)
	if err != nil {
		panic(err)
	}
	return AtClock(t.Clock())
}

func HourSchedule(hour ...int) Schedule {
	var s multiSched
	for _, hour := range hour {
		s = append(s, AtHour(hour))
	}
	return s
}

func MinuteSchedule(min ...int) Schedule {
	var s multiSched
	for _, min := range min {
		s = append(s, AtMinute(min))
	}
	return s
}

func SecondSchedule(sec ...int) Schedule {
	var s multiSched
	for _, sec := range sec {
		s = append(s, AtSecond(sec))
	}
	return s
}

func (c *Clock) Hour(hour int) *Clock {
	if hour > 23 || hour < -1 {
		panic(fmt.Sprintln("invalid hour", hour))
	}
	c.hour = hour
	return c
}

func (c *Clock) Minute(min int) *Clock {
	if min > 59 || min < -1 {
		panic(fmt.Sprintln("invalid minute", min))
	}
	c.min = min
	return c
}

func (c *Clock) Second(sec int) *Clock {
	if sec > 59 || sec < -1 {
		panic(fmt.Sprintln("invalid second", sec))
	}
	c.sec = sec
	return c
}

func (c Clock) IsMatched(t time.Time) bool {
	hour, min, sec := t.Clock()
	return (c.hour == -1 || c.hour == hour) &&
		(c.min == -1 || c.min == min) &&
		(c.sec == -1 || c.sec == sec)
}

func (c Clock) String() string {
	if c.hour == -1 {
		c.hour = 0
	}
	if c.min == -1 {
		c.min = 0
	}
	if c.sec == -1 {
		c.sec = 0
	}
	return fmt.Sprintf("%d:%02d:%02d", c.hour, c.min, c.sec)
}

func (c Clock) Time() time.Time {
	t, _ := time.Parse("15:04:05", c.String())
	return t
}

type clockSched struct {
	start, end *Clock
	d          time.Duration
}

func ClockSchedule(start, end *Clock, d time.Duration) Schedule {
	if d < time.Second || d%time.Second != 0 {
		panic("the minimum duration is one second and must be a multiple of seconds")
	}
	return clockSched{start, end, d}
}

func (s clockSched) IsMatched(t time.Time) bool {
	start, end, t := s.start.Time(), s.end.Time(), AtClock(t.Clock()).Time()
	return (start.Equal(t) || start.Before(t) && end.After(t) || end.Equal(t)) && t.Sub(start)%s.d == 0
}
