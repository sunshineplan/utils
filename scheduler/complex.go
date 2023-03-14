package scheduler

import "time"

var (
	_ Schedule = complexSched(nil)

	_ complexSched = multiSched{}
	_ complexSched = condSched{}
)

type complexSched interface {
	IsMatched(time.Time) bool

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
