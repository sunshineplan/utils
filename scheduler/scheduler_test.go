package scheduler

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler(t *testing.T) {
	now := time.Now()
	s := NewScheduler().At(TimeSchedule(now.Add(time.Second)))
	defer s.Stop()
	var n int32
	if err := s.Do(func(_ time.Time) { atomic.AddInt32(&n, 1) }); err != nil {
		t.Fatal(err)
	}
	if n := atomic.LoadInt32(&n); n != 0 {
		t.Errorf("expected 0; got %d", n)
	}
	time.Sleep(1500 * time.Millisecond)
	if n := atomic.LoadInt32(&n); n != 1 {
		t.Errorf("expected 1; got %d", n)
	}
}

func TestClock(t *testing.T) {
	s := NewScheduler().At(ScheduleFromString(AtClock(time.Now().Add(time.Second).Clock()).String()))
	defer s.Stop()
	var n int32
	if err := s.Do(func(_ time.Time) { atomic.AddInt32(&n, 1) }); err != nil {
		t.Fatal(err)
	}
	if n := atomic.LoadInt32(&n); n != 0 {
		t.Errorf("expected 0; got %d", n)
	}
	time.Sleep(1500 * time.Millisecond)
	if n := atomic.LoadInt32(&n); n != 1 {
		t.Errorf("expected 1; got %d", n)
	}
}

func TestOnce(t *testing.T) {
	now := time.Now()
	s := NewScheduler().At(TimeSchedule(now.Add(time.Second), now.Add(2*time.Second)))
	defer s.Stop()
	var n int32
	done := s.Run(func(_ time.Time) { atomic.AddInt32(&n, 1) }).Once()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
	time.Sleep(1500 * time.Millisecond)
	if n := atomic.LoadInt32(&n); n != 1 {
		t.Errorf("expected 1; got %d", n)
	}
}

func TestTicker(t *testing.T) {
	s := NewScheduler().At(Every(time.Second))
	defer s.Stop()
	var n int32
	if err := s.Run(func(_ time.Time) { atomic.AddInt32(&n, 1) }).Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2500 * time.Millisecond)
	if n := atomic.LoadInt32(&n); n != 2 {
		t.Errorf("expected 2; got %d", n)
	}
}
