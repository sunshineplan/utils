package scheduler

import (
	"testing"
	"time"
)

func TestCondition(t *testing.T) {
	now := time.Now()
	hour, min, sec := now.Clock()
	s := ConditionSchedule(AtHour(hour).Minute(-1).Second(-1), AtMinute(min).Second(-1), AtSecond(sec))
	if !s.IsMatched(now) {
		t.Error("expected true: got false")
	}
	if s.IsMatched(now.Add(time.Second)) {
		t.Error("expected false: got true")
	}
	if s.IsMatched(AtHour(hour).Time()) {
		t.Error("expected false: got true")
	}
}
