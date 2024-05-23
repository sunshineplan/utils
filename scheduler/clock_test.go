package scheduler

import (
	"testing"
	"time"
)

func TestClock(t *testing.T) {
	now := time.Now()
	s := ScheduleFromString(AtClock(now.Clock()).String())
	if !s.IsMatched(now) {
		t.Error("expected true: got false")
	}
	if s.IsMatched(now.Add(2 * time.Second)) {
		t.Error("expected false: got true")
	}
	if s.IsMatched(now.Add(-2 * time.Second)) {
		t.Error("expected false: got true")
	}
	s = ScheduleFromString("7:00")
	if res := s.TickerDuration(); res != 24*time.Hour {
		t.Errorf("expected 24h; got %v", res)
	}
	if res := s.Next(AtClock(6, 0, 0).Time()).Format("15:04:05"); res != "07:00:00" {
		t.Errorf("expected 07:00:00; got %q", res)
	}
}

func TestClockNext1(t *testing.T) {
	ct := AtClock(12, 30, 30).Time()
	for _, testcase := range []struct {
		clock    *Clock
		time     string
		expected time.Duration
	}{
		{AtClock(0, 0, 0), "00:00:00", 11*time.Hour + 29*time.Minute + 30*time.Second},
		{AtClock(-1, -1, -1), "12:30:30", 0},
		{AtClock(-1, -1, 30), "12:30:30", 0},
		{AtClock(-1, 30, -1), "12:30:30", 0},
		{AtClock(12, -1, -1), "12:30:30", 0},
		{AtClock(12, -1, 30), "12:30:30", 0},
		{AtClock(-1, -1, 15), "12:31:15", 45 * time.Second},
		{AtClock(-1, -1, 45), "12:30:45", 15 * time.Second},
		{AtClock(-1, 15, -1), "13:15:00", 44*time.Minute + 30*time.Second},
		{AtClock(-1, 45, -1), "12:45:00", 14*time.Minute + 30*time.Second},
		{AtClock(6, -1, -1), "06:00:00", 17*time.Hour + 29*time.Minute + 30*time.Second}, // +1
		{AtClock(18, -1, -1), "18:00:00", 5*time.Hour + 29*time.Minute + 30*time.Second},
		{AtClock(6, -1, 15), "06:00:15", 17*time.Hour + 29*time.Minute + 45*time.Second}, // +1
		{AtClock(18, -1, 45), "18:00:45", 5*time.Hour + 30*time.Minute + 15*time.Second},
	} {
		next := testcase.clock.Next(ct)
		if res := next.Format("15:04:05"); res != testcase.time {
			t.Errorf("%s expected %q; got %q", testcase.clock, testcase.time, res)
		}
		if res := next.Sub(ct); res != testcase.expected {
			t.Errorf("%s expected %v; got %v", testcase.clock, testcase.expected, res)
		}
	}
}

func TestClockNext2(t *testing.T) {
	for _, testcase := range []struct {
		clock *Clock
		t     time.Time
		s     string
		d     time.Duration
	}{
		{AtClock(0, 0, -1), AtClock(0, 0, 59).Time(), "00:00:59", 0},
		{AtClock(0, -1, 0), AtClock(0, 59, 59).Time(), "00:00:00", 23*time.Hour + time.Second},
		{AtClock(-1, 0, 0), AtClock(23, 59, 0).Time(), "00:00:00", time.Minute},
		{AtClock(0, -1, 0), AtClock(0, 0, 59).Time(), "00:01:00", time.Second},
		{AtClock(0, 0, -1), AtClock(23, 59, 59).Time(), "00:00:00", time.Second},
	} {
		next := testcase.clock.Next(testcase.t)
		if res := next.Format("15:04:05"); res != testcase.s {
			t.Errorf("%s expected %q; got %q", testcase.clock, testcase.s, res)
		}
		if res := next.Sub(testcase.t); res != testcase.d {
			t.Errorf("%s expected %v; got %v", testcase.clock, testcase.d, res)
		}
	}
}

func TestClockTickerDuration(t *testing.T) {
	for _, testcase := range []struct {
		clock    *Clock
		expected time.Duration
	}{
		{AtClock(1, 0, 0), 24 * time.Hour},
		{AtClock(-1, 2, 0), time.Hour},
		{AtClock(0, -1, 3), time.Minute},
		{AtClock(4, 0, -1), time.Second},
		{AtClock(-1, 5, -1), time.Second},
		{AtClock(-1, -1, -1), time.Second},
	} {
		if res := testcase.clock.TickerDuration(); res != testcase.expected {
			t.Errorf("%s expected %v; got %v", testcase.clock, testcase.expected, res)
		}
	}
}

func TestClockSchedule(t *testing.T) {
	s := ClockSchedule(AtHour(9).Minute(30), AtHour(15), 2*time.Minute)
	for _, testcase := range []struct {
		clock    *Clock
		time     string
		expected bool
	}{
		{AtClock(9, 30, 0), "09:30:00", true},
		{AtClock(15, 0, 0), "15:00:00", true},
		{AtClock(13, 0, 0), "13:00:00", true},
		{ClockFromString("9:29"), "09:30:00", false},
		{ClockFromString("9:31"), "09:32:00", false},
		{ClockFromString("16:00:00"), "09:30:00", false},
	} {
		if res := s.Next(testcase.clock.Time()).Format("15:04:05"); res != testcase.time {
			t.Errorf("%s expected %q; got %q", testcase.clock, testcase.time, res)
		}
		if res := s.IsMatched(testcase.clock.Time()); res != testcase.expected {
			t.Errorf("%s expected %v; got %v", testcase.clock, testcase.expected, res)
		}
	}
}

func TestClockScheduleNext(t *testing.T) {
	s := ClockSchedule(ClockFromString("6:00"), ClockFromString("22:00"), time.Hour)
	if res := s.Next(AtClock(21, 59, 0).Time()).Format("15:04:05"); res != "22:00:00" {
		t.Errorf("expected 22:00:00; got %q", res)
	}
}
