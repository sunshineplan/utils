package scheduler

import (
	"testing"
	"time"
)

func TestClock(t *testing.T) {
	now := time.Now()
	s := ScheduleFromString(AtClock(now.Add(time.Second).Clock()).String())
	if s.IsMatched(now) {
		t.Error("expected false: got true")
	}
	if !s.IsMatched(now.Add(time.Second)) {
		t.Error("expected true: got false")
	}
	if s.IsMatched(now.Add(2 * time.Second)) {
		t.Error("expected false: got true")
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
	s := ClockSchedule(AtHour(9).Minute(30), AtHour(15), time.Second)
	for _, testcase := range []struct {
		clock    *Clock
		expected bool
	}{
		{AtClock(9, 30, 0), true},
		{AtClock(15, 0, 0), true},
		{AtClock(13, 0, 0), true},
		{ClockFromString("9:29"), false},
		{ClockFromString("9:31"), true},
		{ClockFromString("16:00:00"), false},
	} {
		if res := s.IsMatched(testcase.clock.Time()); res != testcase.expected {
			t.Errorf("%s expected %v; got %v", testcase.clock, testcase.expected, res)
		}
	}
}
