package scheduler

import "time"

type Schedule interface {
	IsMatched(time.Time) bool
}

var (
	_ Schedule = sched{}
	_ Schedule = weekSched{}
	_ Schedule = weekdaySched{}
	_ Schedule = tickerSched{}
)

var datetimeLayout = []string{
	"2006-01-02",
	"2006-01-02 15:04",
	"2006-01-02 15:04:05",
}

func ScheduleFromString(str ...string) Schedule {
	var s multiSched
	for _, str := range str {
		if _, err := parseTime(str, clockLayout); err == nil {
			s = append(s, ClockFromString(str))
			continue
		}

		t, err := parseTime(str, datetimeLayout)
		if err != nil {
			panic(err)
		}
		s = append(s, TimeSchedule(t))
	}
	return s
}

type sched struct {
	year  int
	month time.Month
	day   int
	clock *Clock
}

func NewSchedule(year int, month time.Month, day int, clock *Clock) Schedule {
	return sched{year, month, day, clock}
}

func TimeSchedule(t ...time.Time) Schedule {
	var s multiSched
	for _, t := range t {
		year, month, day := t.Date()
		s = append(s, sched{year, month, day, AtClock(t.Clock())})
	}
	return s
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

func ISOWeekSchedule(year int, week int, weekday *time.Weekday, clock *Clock) Schedule {
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

func WeekdaySchedule(year int, month time.Month, weekday *time.Weekday, clock *Clock) Schedule {
	return weekdaySched{year, month, weekday, clock}
}

func ptrWeekday(weekday time.Weekday) *time.Weekday {
	return &weekday
}

var Workdays = MultiSchedule(
	WeekdaySchedule(0, 0, ptrWeekday(time.Monday), FullClock),
	WeekdaySchedule(0, 0, ptrWeekday(time.Tuesday), FullClock),
	WeekdaySchedule(0, 0, ptrWeekday(time.Wednesday), FullClock),
	WeekdaySchedule(0, 0, ptrWeekday(time.Thursday), FullClock),
	WeekdaySchedule(0, 0, ptrWeekday(time.Friday), FullClock),
)

var Weekends = MultiSchedule(
	WeekdaySchedule(0, 0, ptrWeekday(time.Saturday), FullClock),
	WeekdaySchedule(0, 0, ptrWeekday(time.Sunday), FullClock),
)

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

func Every(d ...time.Duration) Schedule {
	var s multiSched
	for _, d := range d {
		if d < time.Second || d%time.Second != 0 {
			panic("the minimum duration is one second and must be a multiple of seconds")
		}
		s = append(s, &tickerSched{d: d})
	}
	return s
}

func (s tickerSched) IsMatched(t time.Time) bool {
	if s.d == 0 {
		return false
	}
	return t.Truncate(time.Second).Sub(s.start.Truncate(time.Second))%s.d == 0
}
