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

func TestTickerScheduler(t *testing.T) {
	s := NewScheduler().At(Every(2 * time.Second))
	defer s.Stop()
	var a, b atomic.Int32
	if err := s.Do(func(_ Event) {
		log.Println("TickerScheduler", "a", 1)
		a.Add(1)
	}); err != nil {
		t.Fatal(err)
	}
	log.Println("Start", "TickerScheduler", "a")
	if err := s.Do(func(_ Event) {
		log.Println("TickerScheduler", "b", 1)
		b.Add(1)
	}); err != nil {
		t.Fatal(err)
	}
	log.Println("Start", "TickerScheduler", "b")
	time.Sleep(3 * time.Second)
	if a, b := a.Load(), b.Load(); a != 1 || b != 1 {
		t.Errorf("expected 1 1; got %d %d", a, b)
	}
	time.Sleep(4 * time.Second)
	if a, b := a.Load(), b.Load(); a != 3 || b != 3 {
		t.Errorf("expected 3 3; got %d %d", a, b)
	}
}

func TestIgnoreMissed(t *testing.T) {
	s := NewScheduler().At(Every(time.Minute)).SetIgnoreMissed(true)
	defer s.Stop()
	var n atomic.Int32
	if err := s.Run(func(_ Event) {
		log.Print("Event found")
		n.Add(1)
	}).Start(); err != nil {
		t.Fatal(err)
	}
	log.Print("ignore missed")
	log.Print("send missed event")
	s.tc <- Event{Time: time.Time{}, Goal: time.Time{}, Missed: true}
	time.Sleep(time.Second)
	if n := n.Load(); n != 0 {
		t.Errorf("expected 0; got %d", n)
	}
	log.Print("unignore missed")
	log.Print("send missed event")
	s.SetIgnoreMissed(false)
	s.tc <- Event{Time: time.Time{}, Goal: time.Time{}, Missed: true}
	time.Sleep(time.Second)
	if n := n.Load(); n != 1 {
		t.Errorf("expected 1; got %d", n)
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
	time.Sleep(3 * time.Second)
	if n := n.Load(); n != 1 {
		t.Errorf("expected 1; got %d", n)
	}
}
