package scheduler

import (
	"fmt"
	"time"

	"github.com/sunshineplan/utils"
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
	if clock == nil {
		clock = new(Clock)
	}
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
		return s.clock.IsMatched(t)
	}
	return false
}

func (s sched) Next(t time.Time) (next time.Time) {
	year, month, day := t.Date()
	if s.year != 0 {
		year = s.year
	}
	if s.month != 0 {
		month = s.month
	}
	if s.day != 0 {
		day = s.day
	}
	hour, min, sec := s.clock.Next(t).Clock()
	switch next = time.Date(year, month, day, hour, min, sec, 0, t.Location()); t.Truncate(time.Second).Compare(next) {
	case 1:
		if !s.clock.sec {
			sec = 0
		}
		if !s.clock.min {
			min = 0
		}
		if !s.clock.hour {
			hour = 0
		}
		next = time.Date(year, month, day, hour, min, sec, 0, t.Location())
		if s.day == 0 {
			if next = next.AddDate(0, 0, 1); (next.Month() != month && s.month != 0) ||
				(next.Year() != year && s.year != 0) ||
				t.After(next) {
				next = time.Date(year, month, 1, hour, min, sec, 0, t.Location())
			} else {
				break
			}
		}
		if s.month == 0 {
			if next = next.AddDate(0, 1, 0); next.Year() != year && s.year != 0 || t.After(next) {
				next = time.Date(year, 1, day, hour, min, sec, 0, t.Location())
			} else {
				break
			}
		}
		if s.year == 0 {
			next = next.AddDate(1, 0, 0)
		}
	case -1:
		if !s.clock.sec {
			sec = 0
		}
		if !s.clock.min {
			min = 0
		}
		if !s.clock.hour {
			hour = 0
		}
		if s.day == 0 {
			day = 1
		}
		if s.month == 0 {
			month = 1
		}
		if s.year == 0 {
			year += 1
		}
		next = time.Date(year, month, day, hour, min, sec, 0, t.Location())
	default:
		return t
	}
	if t.After(next) {
		return time.Time{}
	}
	return
}

func (s sched) TickerDuration() time.Duration {
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
	return fmt.Sprintf("%s/%s/%s %s", year, month, day, s.clock)
}

type weekSched struct {
	year, week int
	weekday    *time.Weekday
	clock      *Clock
}

func ISOWeekSchedule(year int, week int, weekday *time.Weekday, clock *Clock) Schedule {
	if clock == nil {
		clock = new(Clock)
	}
	return weekSched{year, week, weekday, clock}
}

func (s weekSched) IsMatched(t time.Time) bool {
	year, week := t.ISOWeek()
	weekday := t.Weekday()
	if y := t.Year(); y == year {
		if (s.year == 0 || s.year == year) &&
			(s.week == 0 || s.week == week) &&
			(s.weekday == nil || *s.weekday == weekday) {
			return s.clock.IsMatched(t)
		}
	} else if (s.year == 0 || s.year == y) &&
		(s.week == 0 || s.week == week) &&
		(s.weekday == nil || *s.weekday == weekday) {
		return s.clock.IsMatched(t)
	}
	return false
}

func (s weekSched) Next(t time.Time) (next time.Time) {
	if s.week < 0 || s.week > 53 {
		return time.Time{}
	}
	year, week := t.ISOWeek()
	weekday := t.Weekday()
	if s.year != 0 {
		year = s.year
	}
	if s.week != 0 {
		week = s.week
	}
	if s.weekday != nil {
		weekday = *s.weekday
	}
	hour, min, sec := s.clock.Next(t).Clock()
	date := func(year int, week int, weekday time.Weekday, hour, min, sec int) time.Time {
		t := time.Date(year, 1, 1, hour, min, sec, 0, t.Location())
		if wd := t.Weekday(); wd != weekday {
			days := int(weekday - wd)
			if days < 0 {
				days += 7
			}
			t = t.AddDate(0, 0, days)
		}
		y, w := t.ISOWeek()
		if y != year {
			t = t.AddDate(0, 0, 7)
			w = 1
		}
		if week != w {
			t = t.AddDate(0, 0, 7*(week-w))
		}
		if y, w := t.ISOWeek(); y != year || w != week {
			return time.Time{}
		}
		return t
	}
	switch next = date(year, week, weekday, hour, min, sec); t.Truncate(time.Second).Compare(next) {
	case 1:
		if !s.clock.sec {
			sec = 0
		}
		if !s.clock.min {
			min = 0
		}
		if !s.clock.hour {
			hour = 0
		}
		next = date(year, week, weekday, hour, min, sec)
		if s.weekday == nil {
			next = next.AddDate(0, 0, 1)
			if y, w := next.ISOWeek(); (w != week && s.week != 0) ||
				(y != year && s.year != 0) ||
				t.After(next) {
				next = date(year, week, time.Monday, hour, min, sec)
			} else {
				break
			}
		}
		if s.week == 0 {
			next = next.AddDate(0, 0, 7)
			if y, _ := next.ISOWeek(); y != year && s.year != 0 || t.After(next) {
				next = date(year, 1, weekday, hour, min, sec)
			} else {
				break
			}
		}
		if s.year == 0 {
			next = next.AddDate(0, 0, 52*7)
			for {
				if _, w := next.ISOWeek(); w == week {
					break
				}
				next = next.AddDate(0, 0, 7)
			}
		}
	case -1:
		if !s.clock.sec {
			sec = 0
		}
		if !s.clock.min {
			min = 0
		}
		if !s.clock.hour {
			hour = 0
		}
		if s.weekday == nil {
			weekday = time.Monday
		}
		if s.week == 0 {
			week = 1
		}
		if s.year == 0 {
			year += 1
		}
		next = date(year, week, weekday, hour, min, sec)
	default:
		return t
	}
	if t.After(next) {
		return time.Time{}
	}
	return
}

func (s weekSched) TickerDuration() time.Duration {
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
	return fmt.Sprintf("%s/ISOWeek:%s/Weekday:%s %s", year, week, weekday, s.clock)
}

type weekdaySched struct {
	year    int
	month   time.Month
	weekday *time.Weekday
	clock   *Clock
}

func WeekdaySchedule(year int, month time.Month, weekday *time.Weekday, clock *Clock) Schedule {
	if clock == nil {
		clock = new(Clock)
	}
	return weekdaySched{year, month, weekday, clock}
}

func Weekday(weekday ...time.Weekday) Schedule {
	var s multiSched
	for _, weekday := range weekday {
		s = append(s, WeekdaySchedule(0, 0, utils.Ptr(weekday), FullClock()))
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
		return s.clock.IsMatched(t)
	}
	return false
}

func (s weekdaySched) Next(t time.Time) time.Time {
	return s.clock.Next(t)
}

func (s weekdaySched) TickerDuration() time.Duration {
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
