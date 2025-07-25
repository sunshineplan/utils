package scheduler

import "testing"

func TestMultiSchedule(t *testing.T) {
	s := MultiSchedule(AtHour(3), AtMinute(4), AtSecond(5))
	for _, testcase := range []struct {
		clock    *Clock
		expected bool
	}{
		{AtClock(3, 0, 0), true},
		{AtClock(0, 4, 0), true},
		{AtClock(0, 0, 5), true},
		{AtClock(4, 5, 3), false},
	} {
		if res := s.IsMatched(testcase.clock.Time()); res != testcase.expected {
			t.Errorf("%s expected %v; got %v", testcase.clock, testcase.expected, res)
		}
	}
}

func TestMultiScheduleNext(t *testing.T) {
	s := MultiSchedule(AtHour(3), AtMinute(4), AtSecond(5))
	for _, testcase := range []struct {
		t    string
		next string
	}{
		{"2000/01/01 00:00:00", "2000/01/01 00:00:05"},
		{"2000/01/01 00:00:05", "2000/01/01 00:01:05"},
		{"2000/01/01 00:03:05", "2000/01/01 00:04:00"},
		{"2000/01/01 00:04:00", "2000/01/01 00:04:05"},
		{"2000/01/01 02:59:05", "2000/01/01 03:00:00"},
		{"2000/01/01 03:00:00", "2000/01/01 03:00:05"},
	} {
		if res := s.Next(parse(testcase.t)).Format(format); res != testcase.next {
			t.Errorf("%s expected %v; got %v", testcase.t, testcase.next, res)
		}
	}
}

func TestConditionSchedule(t *testing.T) {
	s := ConditionSchedule(Weekdays, MultiSchedule(AtClock(9, 30, 0), AtHour(15)))
	for _, testcase := range []struct {
		clock    *Clock
		expected bool
	}{
		{AtClock(9, 0, 0), false},
		{AtClock(9, 30, 0), true},
		{AtClock(15, 0, 0), true},
		{AtClock(15, 30, 0), false},
	} {
		if res := s.IsMatched(testcase.clock.Time()); res != testcase.expected {
			t.Errorf("%s expected %v; got %v", testcase.clock, testcase.expected, res)
		}
	}
}

func TestConditionScheduleNext(t *testing.T) {
	s := ConditionSchedule(Weekdays, MultiSchedule(AtClock(9, 30, 0), AtHour(15)))
	for _, testcase := range []struct {
		t    string
		next string
	}{
		{"2000/01/01 00:00:00", "2000/01/03 09:30:00"},
		{"2000/01/03 09:30:00", "2000/01/03 15:00:00"},
		{"2000/01/03 15:00:00", "2000/01/04 09:30:00"},
		{"2000/01/07 15:00:00", "2000/01/10 09:30:00"},
	} {
		if res := s.Next(parse(testcase.t)).Format(format); res != testcase.next {
			t.Errorf("%s expected %v; got %v", testcase.t, testcase.next, res)
		}
	}
}
