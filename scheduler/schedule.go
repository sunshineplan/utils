package scheduler

import (
	"fmt"
	"time"

	"github.com/sunshineplan/utils/clock"
)

type Schedule interface {
	IsMatched(time.Time) bool
	Next(time.Time) time.Time
	TickerDuration() time.Duration
	String() string
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
		if _, err := clock.Parse(str); err == nil {
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
			s.clock = &Clock{}
		}
		return s.clock.IsMatched(t)
	}
	return false
}

func (s sched) Next(t time.Time) time.Time {
	if s.clock == nil {
		s.clock = &Clock{}
	}
	return s.clock.Next(t)
}

func (s sched) TickerDuration() time.Duration {
	if s.clock == nil {
		return 24 * time.Hour
	}
	return s.clock.TickerDuration()
}

func (s sched) String() string {
	var year, month, day string
	if s.year == 0 {
		year = "----"
	} else {
		year = fmt.Sprint(s.year)
	}
	if s.month == 0 {
		month = "--"
	} else {
		month = fmt.Sprintf("%02d", s.month)
	}
	if s.day == 0 {
		day = "--"
	} else {
		day = fmt.Sprintf("%02d", s.day)
	}
	if s.clock == nil {
		s.clock = &Clock{}
	}
	return fmt.Sprintf("%s/%s/%s %s", year, month, day, s.clock)
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
			s.clock = &Clock{}
		}
		return s.clock.IsMatched(t)
	}
	return false
}

func (s weekSched) Next(t time.Time) time.Time {
	if s.clock == nil {
		s.clock = &Clock{}
	}
	return s.clock.Next(t)
}

func (s weekSched) TickerDuration() time.Duration {
	if s.clock == nil {
		return 24 * time.Hour
	}
	return s.clock.TickerDuration()
}

func (s weekSched) String() string {
	var year, week, weekday string
	if s.year == 0 {
		year = "----"
	} else {
		year = fmt.Sprint(s.year)
	}
	if s.week == 0 {
		week = "--"
	} else {
		week = fmt.Sprintf("%02d", s.week)
	}
	if s.weekday == nil {
		weekday = "----"
	} else {
		weekday = fmt.Sprint(s.weekday)
	}
	if s.clock == nil {
		s.clock = &Clock{}
	}
	return fmt.Sprintf("%s/ISOWeek:%s/Weekday:%s %s", year, week, weekday, s.clock)
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

func Weekday(weekday ...time.Weekday) Schedule {
	var s multiSched
	for _, weekday := range weekday {
		s = append(s, WeekdaySchedule(0, 0, ptrWeekday(weekday), FullClock()))
	}
	return s
}

var (
	Weekdays = Weekday(time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday)
	Weekends = Weekday(time.Saturday, time.Sunday)
)

func (s weekdaySched) IsMatched(t time.Time) bool {
	year, month, _ := t.Date()
	weekday := t.Weekday()
	if (s.year == 0 || s.year == year) &&
		(s.month == 0 || s.month == month) &&
		(s.weekday == nil || *s.weekday == weekday) {
		if s.clock == nil {
			s.clock = &Clock{}
		}
		return s.clock.IsMatched(t)
	}
	return false
}

func (s weekdaySched) Next(t time.Time) time.Time {
	if s.clock == nil {
		s.clock = &Clock{}
	}
	return s.clock.Next(t)
}

func (s weekdaySched) TickerDuration() time.Duration {
	if s.clock == nil {
		return 24 * time.Hour
	}
	return s.clock.TickerDuration()
}

func (s weekdaySched) String() string {
	var year, month, weekday string
	if s.year == 0 {
		year = "----"
	} else {
		year = fmt.Sprint(s.year)
	}
	if s.month == 0 {
		month = "--"
	} else {
		month = fmt.Sprintf("%02d", s.month)
	}
	if s.weekday == nil {
		weekday = "----"
	} else {
		weekday = fmt.Sprint(s.weekday)
	}
	if s.clock == nil {
		s.clock = &Clock{}
	}
	return fmt.Sprintf("%s/%s/Weekday:%s %s", year, month, weekday, s.clock)
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

func (s *tickerSched) init(t time.Time) {
	s.start = t.Truncate(time.Second).Add(time.Second)
}

func (s tickerSched) IsMatched(t time.Time) bool {
	if s.d == 0 {
		return false
	}
	return t.Truncate(time.Second).Sub(s.start)%s.d == 0
}

func (s tickerSched) Next(t time.Time) time.Time {
	if d := t.Sub(s.start); d > 0 {
		return t.Add(d % s.d)
	}
	return s.start
}

func (s tickerSched) TickerDuration() time.Duration {
	return s.d
}

func (s tickerSched) String() string {
	return fmt.Sprint("Every ", s.d)
}
