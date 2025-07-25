package scheduler

import (
	"log"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler(t *testing.T) {
	now := time.Now()
	s := NewScheduler().At(TimeSchedule(now.Add(2 * time.Second)))
	defer s.Stop()
	var n atomic.Int32
	if err := s.Run(func(_ Event) {
		log.Println("Scheduler", 1)
		n.Add(1)
	}).Start(); err != nil {
		t.Fatal(err)
	}
	log.Println("Start", "Scheduler")
	if n := n.Load(); n != 0 {
		t.Errorf("expected 0; got %d", n)
	}
	time.Sleep(3 * time.Second)
	if n := n.Load(); n != 1 {
		t.Errorf("expected 1; got %d", n)
	}
}

func TestTickerScheduler1(t *testing.T) {
	s := NewScheduler().At(Every(time.Second))
	defer s.Stop()
	var a, b atomic.Int32
	if err := s.Do(func(_ Event) {
		log.Println("TickerScheduler1", "a", 1)
		a.Add(1)
	}); err != nil {
		t.Fatal(err)
	}
	log.Println("Start", "TickerScheduler1", "a")
	if err := s.Do(func(_ Event) {
		log.Println("TickerScheduler1", "b", 1)
		b.Add(1)
	}); err != nil {
		t.Fatal(err)
	}
	log.Println("Start", "TickerScheduler1", "b")
	time.Sleep(1500 * time.Millisecond)
	if a, b := a.Load(), b.Load(); a != 1 || b != 1 {
		t.Errorf("expected 1 1; got %d %d", a, b)
	}
	time.Sleep(1600 * time.Millisecond)
	if a, b := a.Load(), b.Load(); a != 3 || b != 3 {
		t.Errorf("expected 3 3; got %d %d", a, b)
	}
}

func TestOnce(t *testing.T) {
	now := time.Now()
	s := NewScheduler().At(TimeSchedule(now.Add(time.Second), now.Add(2*time.Second)))
	defer s.Stop()
	var n atomic.Int32
	done := s.Run(func(_ Event) {
		log.Println("Once", 1)
		n.Add(1)
	}).Once()
	log.Println("Start", "Once")
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
