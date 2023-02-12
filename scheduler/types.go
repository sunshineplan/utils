package scheduler

import "time"

type Time interface {
	IsMatched(time.Time) bool
}

var (
	_ Time = sched{}
	_ Time = Clock{}
	_ Time = weekSched{}
	_ Time = weekdaySched{}
	_ Time = tickerSched{}
)

var (
	datetimeLayout = []string{
		"2006-01-02",
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
	}
	clockLayout = []string{
		"15:04",
		"15:04:05",
	}
)

func ScheduleFromString(s string) Time {
	if t, err := parseTime(s, clockLayout); err == nil {
		return AtClock(t.Clock())
	}

	t, err := parseTime(s, datetimeLayout)
	if err != nil {
		panic(err)
	}
	return TimeSchedule(t)
}

type sched struct {
	year  int
	month time.Month
	day   int
	clock *Clock
}

func Schedule(year int, month time.Month, day int, clock *Clock) Time {
	return sched{year, month, day, clock}
}

func TimeSchedule(t time.Time) Time {
	year, month, day := t.Date()
	return sched{year, month, day, AtClock(t.Clock())}
}

func (s sched) IsMatched(t time.Time) bool {
	year, month, day := t.Date()
	if (s.year == 0 || s.year == year) &&
		(s.month == 0 || s.month == month) &&
		(s.day == 0 || s.day == day) {
		if s.clock == nil {
			hour, min, sec := t.Clock()
			return hour == 0 && min == 0 && sec == 0
		}
		return s.clock.IsMatched(t)
	}
	return false
}

type weekSched struct {
	year, week int
	weekday    *time.Weekday
	clock      *Clock
}

func ISOWeekSchedule(year int, week int, weekday *time.Weekday, clock *Clock) Time {
	return weekSched{year, week, weekday, clock}
}

func (s weekSched) IsMatched(t time.Time) bool {
	year, week := t.ISOWeek()
	weekday := t.Weekday()
	if (s.year == 0 || s.year == year) &&
		(s.week == 0 || s.week == week) &&
		(s.weekday == nil || *s.weekday == weekday) {
		if s.clock == nil {
			hour, min, sec := t.Clock()
			return hour == 0 && min == 0 && sec == 0
		}
		return s.clock.IsMatched(t)
	}
	return false
}

type weekdaySched struct {
	year    int
	month   time.Month
	weekday *time.Weekday
	clock   *Clock
}

func WeekdaySchedule(year int, month time.Month, weekday *time.Weekday, clock *Clock) Time {
	return weekdaySched{year, month, weekday, clock}
}

func (s weekdaySched) IsMatched(t time.Time) bool {
	year, month, _ := t.Date()
	weekday := t.Weekday()
	if (s.year == 0 || s.year == year) &&
		(s.month == 0 || s.month == month) &&
		(s.weekday == nil || *s.weekday == weekday) {
		if s.clock == nil {
			hour, min, sec := t.Clock()
			return hour == 0 && min == 0 && sec == 0
		}
		return s.clock.IsMatched(t)
	}
	return false
}

type tickerSched struct {
	d     time.Duration
	start time.Time
}

func Every(d time.Duration) Time {
	if d < time.Second || d%time.Second != 0 {
		panic("")
	}
	return &tickerSched{d: d}
}

func (s tickerSched) IsMatched(t time.Time) bool {
	if s.d == 0 {
		return false
	}
	return t.Truncate(time.Second).Sub(s.start.Truncate(time.Second))%s.d == 0
}
