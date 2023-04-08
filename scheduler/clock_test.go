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

func TestClockFirst(t *testing.T) {
	for _, testcase := range []struct {
		clock    *Clock
		expected time.Duration
	}{
		{AtClock(0, 0, 0), 11*time.Hour + 29*time.Minute + 30*time.Second}, // 00:00:00
		{AtClock(-1, -1, -1), 0},                                             // 12:30:30
		{AtClock(-1, -1, 30), 0},                                             // 12:30:30
		{AtClock(-1, 30, -1), 0},                                             // 12:30:30
		{AtClock(12, -1, -1), 0},                                             // 12:30:30
		{AtClock(12, -1, 30), 0},                                             // 12:30:30
		{AtClock(-1, -1, 15), 45 * time.Second},                              // 12:31:15
		{AtClock(-1, -1, 45), 15 * time.Second},                              // 12:30:45
		{AtClock(-1, 15, -1), 44*time.Minute + 30*time.Second},               // 13:15:00
		{AtClock(-1, 45, -1), 14*time.Minute + 30*time.Second},               // 12:45:00
		{AtClock(6, -1, -1), 17*time.Hour + 29*time.Minute + 30*time.Second}, // 06:00:00+1
		{AtClock(18, -1, -1), 5*time.Hour + 29*time.Minute + 30*time.Second}, // 18:00:00
		{AtClock(6, -1, 15), 17*time.Hour + 29*time.Minute + 45*time.Second}, // 06:00:15+1
		{AtClock(18, -1, 45), 5*time.Hour + 30*time.Minute + 15*time.Second}, // 18:00:45
	} {
		if res := testcase.clock.First(AtClock(12, 30, 30).Time()); res != testcase.expected {
			t.Errorf("%s expected %v; got %v", testcase.clock, testcase.expected, res)
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
		duration time.Duration
		expected bool
	}{
		{AtClock(9, 30, 0), 0, true},
		{AtClock(15, 0, 0), 0, true},
		{AtClock(13, 0, 0), 0, true},
		{ClockFromString("9:29"), time.Minute, false},
		{ClockFromString("9:31"), time.Minute, false},
		{ClockFromString("16:00:00"), 17*time.Hour + 30*time.Minute, false},
	} {
		if res := s.First(testcase.clock.Time()); res != testcase.duration {
			t.Errorf("%s expected %v; got %v", testcase.clock, testcase.expected, res)
		}
		if res := s.IsMatched(testcase.clock.Time()); res != testcase.expected {
			t.Errorf("%s expected %v; got %v", testcase.clock, testcase.expected, res)
		}
	}
}
