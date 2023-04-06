package scheduler

import "time"

var (
	_ Schedule = complexSched(nil)

	_ complexSched = multiSched{}
	_ complexSched = condSched{}
)

type complexSched interface {
	IsMatched(time.Time) bool
	First(time.Time) time.Duration
	TickerDuration() time.Duration

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
			i.start = t
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

func (s multiSched) First(t time.Time) time.Duration {
	var res time.Duration
	for i, sched := range s {
		if i == 0 {
			res = sched.First(t)
		} else if d := sched.First(t); d < res {
			res = d
		}
	}
	return res
}

func (s multiSched) TickerDuration() time.Duration {
	if len(s) == 1 {
		return s[0].TickerDuration()
	}
	return time.Second
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

func (s condSched) First(t time.Time) time.Duration {
	if len(s) == 1 {
		return s[0].First(t)
	}
	return 0
}

func (s condSched) TickerDuration() time.Duration {
	if len(s) == 1 {
		return s[0].TickerDuration()
	}
	return time.Second
}
