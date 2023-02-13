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

func TestClockSchedule(t *testing.T) {
	s := ClockSchedule(AtHour(9).Minute(30), AtHour(15), time.Second)
	for _, testcase := range []struct {
		clock    *Clock
		expected bool
	}{
		{AtClock(9, 30, 0), true},
		{AtClock(15, 0, 0), true},
		{AtClock(13, 0, 0), true},
		{AtClock(9, 29, 0), false},
		{AtClock(9, 31, 0), true},
		{AtClock(16, 0, 0), false},
	} {
		if res := s.IsMatched(testcase.clock.Time()); res != testcase.expected {
			t.Errorf("%s expected %v; got %v", testcase.clock, testcase.expected, res)
		}
	}
}
