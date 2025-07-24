package scheduler

import (
	"testing"
	"time"
)

func TestMultiSchedule(t *testing.T) {
	s := MultiSchedule(AtHour(3), AtMinute(4), AtSecond(5))
	if res := s.Next(time.Time{}).Format("15:04:05"); res != "00:00:05" {
		t.Fatalf("expected 00:00:05: got %q", res)
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

func TestConditionSchedule(t *testing.T) {
	s := ConditionSchedule(Weekdays, MultiSchedule(AtClock(9, 30, 0), AtHour(15)))
	if res := s.Next(time.Time{}).Format("15:04:05"); res != "09:30:00" {
		t.Fatalf("expected 09:30:00: got %q", res)
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
