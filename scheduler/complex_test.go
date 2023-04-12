package scheduler

import (
	"testing"
	"time"
)

func TestMultiSchedule(t *testing.T) {
	s := MultiSchedule(AtHour(3), AtMinute(4), AtSecond(5))
	if d := s.TickerDuration(); d != time.Minute {
		t.Fatalf("expected 1m: got %s", d)
	}
	if d := s.First(time.Date(2006, 1, 2, 0, 0, 0, 0, time.Local)); d != 5*time.Second {
		t.Fatalf("expected 5s: got %s", d)
	}
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

func TestMultiScheduleTickerDuration(t *testing.T) {
	for _, testcase := range []struct {
		schedule Schedule
		expected time.Duration
	}{
		{MultiSchedule(AtHour(3), AtMinute(4), AtSecond(5)), time.Minute},
		{HourSchedule(9, 17, 23), time.Hour},
		{MinuteSchedule(0, 59), time.Minute},
	} {
		if res := testcase.schedule.TickerDuration(); res != testcase.expected {
			t.Errorf("expected %v; got %v", testcase.expected, res)
		}
	}
}

func TestConditionSchedule(t *testing.T) {
	s := ConditionSchedule(Weekends, MultiSchedule(AtClock(9, 30, 0), AtHour(15)))
	if d := s.TickerDuration(); d != time.Second {
		t.Fatalf("expected 1s: got %s", d)
	}
	if d := s.First(time.Date(2006, 1, 2, 3, 4, 5, 0, time.Local)); d != time.Second {
		t.Fatalf("expected 1s: got %s", d)
	}
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
