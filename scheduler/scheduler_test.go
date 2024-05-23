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
	var n atomic.Int32
	if err := s.Run(func(_ time.Time) { n.Add(1) }).Start(); err != nil {
		t.Fatal(err)
	}
	if n := n.Load(); n != 0 {
		t.Errorf("expected 0; got %d", n)
	}
	time.Sleep(1500 * time.Millisecond)
	if n := n.Load(); n != 1 {
		t.Errorf("expected 1; got %d", n)
	}
}

func TestTickerScheduler1(t *testing.T) {
	s := NewScheduler().At(Every(time.Second))
	defer s.Stop()
	var a, b atomic.Int32
	if err := s.Do(func(_ time.Time) { a.Store(1) }); err != nil {
		t.Fatal(err)
	}
	if err := s.Do(func(_ time.Time) { b.Store(1) }); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1500 * time.Millisecond)
	if a, b := a.Load(), b.Load(); a != 1 || b != 1 {
		t.Errorf("expected 1 1; got %d %d", a, b)
	}
}

func TestTickerScheduler2(t *testing.T) {
	var n atomic.Int32
	s := NewScheduler().At(Every(time.Minute)).Run(func(_ time.Time) { n.Add(1) })
	defer s.Stop()
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1500 * time.Millisecond)
	if n := n.Load(); n != 1 {
		t.Errorf("expected 1; got %d", n)
	}

	s.mu.Lock()
	if s.timer.Stop() {
		t.Fatal("expected timer stopped; got running")
	}
	s.mu.Unlock()

	s.notify <- time.Now()
	time.Sleep(500 * time.Millisecond)

	s.mu.Lock()
	if !s.timer.Stop() {
		t.Fatal("expected timer running; got stopped")
	}
	s.mu.Unlock()
}

func TestOnce(t *testing.T) {
	now := time.Now()
	s := NewScheduler().At(TimeSchedule(now.Add(time.Second), now.Add(2*time.Second)))
	defer s.Stop()
	var n atomic.Int32
	done := s.Run(func(_ time.Time) { n.Add(1) }).Once()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2500 * time.Millisecond):
		t.Fatal("timeout")
	}
	time.Sleep(1500 * time.Millisecond)
	if n := n.Load(); n != 1 {
		t.Errorf("expected 1; got %d", n)
	}
}
