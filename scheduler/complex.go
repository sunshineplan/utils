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

type complexSched interface {
	IsMatched(time.Time) bool
	Next(time.Time) time.Time
	TickerDuration() time.Duration
	String() string

	init(t time.Time)
	len() int
}

type complex interface {
	~[]Schedule
}

func initComplexSched[sche complex](s sche, t time.Time) {
	for _, s := range s {
		if i, ok := s.(complexSched); ok {
			i.init(t)
		} else if i, ok := s.(*tickerSched); ok {
			i.init(t)
		}
	}
}

type multiSched []Schedule

func MultiSchedule(schedules ...Schedule) Schedule {
	return multiSched(schedules)
}

func (s multiSched) init(t time.Time) {
	initComplexSched(s, t)
}

func (s multiSched) len() int {
	return len(s)
}

func (s multiSched) IsMatched(t time.Time) bool {
	for _, i := range s {
		if i.IsMatched(t) {
			return true
		}
	}
	return false
}

func (s multiSched) Next(t time.Time) (next time.Time) {
	for i, sched := range s {
		if i == 0 {
			next = sched.Next(t)
		} else if t := sched.Next(t); t.Before(next) {
			next = t
		}
	}
	return
}

func (s multiSched) TickerDuration() time.Duration {
	var downgrade bool
	var res time.Duration
	for _, i := range s {
		d := gcd(res, i.TickerDuration())
		if !downgrade && res == d {
			downgrade = true
		}
		res = d
	}
	if downgrade {
		switch res {
		case 24 * time.Hour:
			return time.Hour
		case time.Hour:
			return time.Minute
		case time.Minute:
			return time.Second
		}
	}
	return res
}

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

type condSched []Schedule

func ConditionSchedule(schedules ...Schedule) Schedule {
	return condSched(schedules)
}

func (s condSched) init(t time.Time) {
	initComplexSched(s, t)
}

func (s condSched) len() int {
	return len(s)
}

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

func (s condSched) Next(t time.Time) time.Time {
	if len(s) == 1 {
		return s[0].Next(t)
	}
	d := s.TickerDuration()
	return t.Add(d).Truncate(d)
}

func (s condSched) TickerDuration() time.Duration {
	var res time.Duration
	for _, i := range s {
		res = gcd(res, i.TickerDuration())
	}
	return res
}

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
