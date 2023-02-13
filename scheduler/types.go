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
	_ Time = multiSched{}
)

var (
	clockLayout = []string{
		"15:04",
		"15:04:05",
	}
	datetimeLayout = []string{
		"2006-01-02",
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
	}
)

func ScheduleFromString(str ...string) Time {
	var s multiSched
	for _, str := range str {
		if t, err := parseTime(str, clockLayout); err == nil {
			s = append(s, AtClock(t.Clock()))
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

func Schedule(year int, month time.Month, day int, clock *Clock) Time {
	return sched{year, month, day, clock}
}

func TimeSchedule(t ...time.Time) Time {
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

func Every(d ...time.Duration) Time {
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

type multiSched []Time

func HourSchedule(hour ...int) Time {
	var s multiSched
	for _, hour := range hour {
		s = append(s, AtHour(hour))
	}
	return s
}

func MinuteSchedule(min ...int) Time {
	var s multiSched
	for _, min := range min {
		s = append(s, AtMinute(min))
	}
	return s
}

func SecondSchedule(sec ...int) Time {
	var s multiSched
	for _, sec := range sec {
		s = append(s, AtSecond(sec))
	}
	return s
}

func ClockRange(start, end *Clock, d time.Duration) Time {
	if d < time.Second || d%time.Second != 0 {
		panic("the minimum duration is one second and must be a multiple of seconds")
	}
	var s multiSched
	for t, end := start.Time(), end.Time(); t.Before(end) || t.Equal(end); t = t.Add(d) {
		s = append(s, AtClock(t.Clock()))
	}
	return s
}

func (s *multiSched) init(t time.Time) {
	for _, s := range *s {
		if i, ok := s.(multiSched); ok {
			i.init(t)
		} else if i, ok := s.(*tickerSched); ok {
			i.start = t
		}
	}
}

func (s multiSched) IsMatched(t time.Time) bool {
	for _, i := range s {
		if i.IsMatched(t) {
			return true
		}
	}
	return false
}
