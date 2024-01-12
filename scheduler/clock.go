package scheduler

import (
	"fmt"
	"time"

	"github.com/sunshineplan/utils/clock"
)

var (
	_ Schedule = Clock{}
	_ Schedule = clockSched{}
)

type Clock struct {
	clock.Clock
	hour, min, sec bool
}

func atClock(hour, min, sec int) clock.Clock {
	if hour == -1 {
		hour = 0
	}
	if min == -1 {
		min = 0
	}
	if sec == -1 {
		sec = 0
	}
	return clock.New(hour, min, sec)
}

func AtClock(hour, min, sec int) *Clock {
	if hour > 23 || hour < -1 ||
		min > 59 || min < -1 ||
		sec > 59 || sec < -1 {
		panic(fmt.Sprintf("invalid clock: hour(%d) min(%d) sec(%d)", hour, min, sec))
	}
	var c Clock
	if hour > -1 {
		c.hour = true
	}
	if min > -1 {
		c.min = true
	}
	if sec > -1 {
		c.sec = true
	}
	c.Clock = atClock(hour, min, sec)
	return &c
}

func FullClock() *Clock { return AtClock(-1, -1, -1) }

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
	c, err := clock.Parse(str)
	if err != nil {
		panic(err)
	}
	return &Clock{c, true, true, true}
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
		panic(fmt.Sprint("invalid hour ", hour))
	}
	if hour > -1 {
		c.hour = true
	} else {
		c.hour = false
	}
	c.Clock = atClock(hour, c.Clock.Minute(), c.Clock.Second())
	return c
}

func (c *Clock) Minute(min int) *Clock {
	if min > 59 || min < -1 {
		panic(fmt.Sprint("invalid minute ", min))
	}
	if min > -1 {
		c.min = true
	} else {
		c.min = false
	}
	c.Clock = atClock(c.Clock.Hour(), min, c.Clock.Second())
	return c
}

func (c *Clock) Second(sec int) *Clock {
	if sec > 59 || sec < -1 {
		panic(fmt.Sprint("invalid second ", sec))
	}
	if sec > -1 {
		c.sec = true
	} else {
		c.sec = false
	}
	c.Clock = atClock(c.Clock.Hour(), c.Clock.Minute(), sec)
	return c
}

func (c Clock) IsMatched(t time.Time) bool {
	hour, min, sec := t.Clock()
	return (!c.hour || c.Clock.Hour() == hour) &&
		(!c.min || c.Clock.Minute() == min) &&
		(!c.sec || c.Clock.Second() == sec)
}

func (c Clock) First(t time.Time) time.Duration {
	u, nc := AtClock(t.Clock()), new(Clock)
	if c.sec {
		nc.Second(c.Clock.Second())
	} else {
		nc.Second(u.Clock.Second())
	}
	if c.min {
		nc.Minute(c.Clock.Minute())
	} else {
		nc.Minute(u.Clock.Minute())
	}
	if c.hour {
		nc.Hour(c.Clock.Hour())
	} else {
		nc.Hour(u.Clock.Hour())
	}

	if nc.Equal(u.Clock) {
		return 0
	}
	if !c.sec {
		nc.Second(0)
	}
	if nc.Before(u.Clock) {
		if !c.min {
			if nc.Clock.Hour() == u.Clock.Hour() {
				return nc.Add(time.Minute).Since(u.Clock)
			}
			nc.Minute(0)
		}
		if !c.hour {
			return nc.Add(time.Hour).Since(u.Clock)
		}
		return nc.Add(24 * time.Hour).Since(u.Clock)
	} else if !c.min && nc.Clock.Hour() != u.Clock.Hour() {
		nc.Minute(0)
	}
	return nc.Since(u.Clock)
}

func (c Clock) TickerDuration() time.Duration {
	if !c.sec {
		return time.Second
	} else if !c.min {
		return time.Minute
	} else if !c.hour {
		return time.Hour
	} else {
		return 24 * time.Hour
	}
}

func (c Clock) String() string {
	var hour, min, sec string
	if !c.hour {
		hour = "--"
	} else {
		hour = fmt.Sprint(c.Clock.Hour())
	}
	if !c.min {
		min = "--"
	} else {
		min = fmt.Sprintf("%02d", c.Clock.Minute())
	}
	if !c.sec {
		sec = "--"
	} else {
		sec = fmt.Sprintf("%02d", c.Clock.Second())
	}
	return fmt.Sprintf("%s:%s:%s", hour, min, sec)
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
	start, end, tc := s.start, s.end, AtClock(t.Clock()).Clock
	return (start.Equal(tc) || start.Before(tc) && end.After(tc) || end.Equal(tc)) && tc.Since(start.Clock)%s.d == 0
}

func (s clockSched) First(t time.Time) time.Duration {
	if s.IsMatched(t) {
		return 0
	}
	start, end, tc := s.start.Clock, s.end.Clock, AtClock(t.Clock()).Clock
	for c := AtClock(t.Clock()); c.Compare(start) != -1 && c.Compare(end) != 1; c.Clock = c.Add(time.Second) {
		if s.IsMatched(c.Time()) {
			return c.Since(tc)
		}
	}
	return s.start.First(t)
}

func (s clockSched) TickerDuration() time.Duration {
	if s.start.Clock.Second() != 0 {
		return time.Second
	} else if s.start.Clock.Minute() != 0 && s.d%time.Minute == 0 {
		return time.Minute
	}
	return s.d
}

func (s clockSched) String() string {
	return fmt.Sprintf("%q-%q(every %s)", s.start, s.end, s.d)
}
