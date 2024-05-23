package scheduler

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrNoFunction     = errors.New("scheduler function is not set")
	ErrNoSchedule     = errors.New("no schedule has been added")
	ErrAlreadyRunning = errors.New("scheduler is already running")
)

type Scheduler struct {
	mu     sync.Mutex
	notify chan time.Time

	timer  *time.Timer
	ticker *time.Ticker
	tc     chan time.Time
	fn     []func(time.Time)

	sched complexSched
	d     time.Duration
	next  time.Time

	ctx    context.Context
	cancel context.CancelFunc
}

func NewScheduler() *Scheduler {
	return &Scheduler{notify: make(chan time.Time, 1), tc: make(chan time.Time, 1)}
}

func (sched *Scheduler) At(schedules ...Schedule) *Scheduler {
	if len(schedules) == 0 {
		panic("no schedules")
	}
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.sched = multiSched(schedules)
	sched.d = sched.sched.TickerDuration()
	return sched
}

func (sched *Scheduler) AtCondition(schedules ...Schedule) *Scheduler {
	if len(schedules) == 0 {
		panic("no schedules")
	}
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.sched = condSched(schedules)
	sched.d = sched.sched.TickerDuration()
	return sched
}

func (sched *Scheduler) Clear() {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.sched = nil
	sched.d = 0
	if sched.ctx != nil && sched.ctx.Err() == nil {
		sched.Stop()
	}
}

func (sched *Scheduler) Run(fn ...func(time.Time)) *Scheduler {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	for _, fn := range fn {
		if fn != nil {
			sched.fn = append(sched.fn, fn)
		}
	}
	return sched
}

func (sched *Scheduler) init() error {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	if len(sched.fn) == 0 {
		return ErrNoFunction
	} else if sched.sched.len() == 0 {
		return ErrNoSchedule
	}
	if sched.ctx == nil || sched.ctx.Err() != nil {
		sched.ctx, sched.cancel = context.WithCancel(context.Background())
		now := time.Now()
		timer := time.NewTimer(now.Truncate(time.Second).Add(1020 * time.Millisecond).Sub(now))
		defer timer.Stop()
		t := <-timer.C
		sched.sched.init(t)
		sched.next = sched.sched.Next(t)
		subscribeNotify(sched.notify)
		sched.newTimer(time.Now())
		return nil
	}
	return ErrAlreadyRunning
}

func (sched *Scheduler) checkMatched(t time.Time) {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	var matched time.Time
	var notify bool
	if sched.sched.IsMatched(t) {
		sched.tc <- t
		matched = t
	} else if sched.sched.TickerDuration() >= time.Minute {
		if minus1s := t.Add(-time.Second); sched.sched.IsMatched(minus1s) {
			sched.tc <- minus1s
			matched = minus1s
			notify = true
		} else if plus1s := t.Add(time.Second); sched.sched.IsMatched(plus1s) {
			sched.tc <- plus1s
			matched = plus1s
			time.Sleep(2 * time.Second)
			notify = true
		}
	}
	if !matched.IsZero() && !notify {
		if sched.next = sched.sched.Next(matched.Truncate(time.Second).Add(time.Second)); sched.next.IsZero() {
			sched.Stop()
			return
		}
	} else if t.After(sched.next) {
		notify = true
	}
	if notify {
		sched.notify <- time.Now()
	}
}

func (sched *Scheduler) newTimer(t time.Time) {
	if sched.next.IsZero() {
		sched.Stop()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	sched.timer = time.AfterFunc(sched.next.Sub(t), func() {
		cancel()
		sched.ticker = time.NewTicker(sched.d)
		go func() {
			for {
				select {
				case t := <-sched.ticker.C:
					sched.checkMatched(t)
				case t := <-sched.notify:
					sched.ticker.Stop()
					sched.mu.Lock()
					defer sched.mu.Unlock()
					sched.next = sched.sched.Next(t)
					sched.newTimer(t)
					return
				case <-sched.ctx.Done():
					sched.ticker.Stop()
					return
				}
			}
		}()
		sched.checkMatched(time.Now())
	})
	go func() {
		for {
			select {
			case t := <-sched.notify:
				sched.mu.Lock()
				if sched.timer.Stop() {
					sched.next = sched.sched.Next(t)
					if sched.next.IsZero() {
						cancel()
						sched.Stop()
						return
					}
					sched.timer.Reset(sched.next.Sub(t))
				}
				sched.mu.Unlock()
			case <-sched.ctx.Done():
				cancel()
				sched.timer.Stop()
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (sched *Scheduler) Start() error {
	if err := sched.init(); err != nil {
		return err
	}
	go func() {
		for {
			select {
			case t := <-sched.tc:
				sched.mu.Lock()
				for _, fn := range sched.fn {
					go fn(t)
				}
				sched.mu.Unlock()
			case <-sched.ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (sched *Scheduler) Stop() {
	sched.cancel()
	unsubscribeNotify(sched.notify)
}

func (sched *Scheduler) immediately(t time.Time) <-chan error {
	sched.mu.Lock()
	defer sched.mu.Unlock()

	var wg sync.WaitGroup
	wg.Add(len(sched.fn))
	for _, fn := range sched.fn {
		go func(f func(time.Time)) {
			defer wg.Done()
			f(t)
		}(fn)
	}
	done := make(chan error)
	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}

func (sched *Scheduler) Immediately() <-chan error {
	return sched.immediately(time.Now())
}

func (sched *Scheduler) Once() <-chan error {
	done := make(chan error)
	go func() {
		if err := sched.init(); err != nil {
			done <- err
			return
		}
		select {
		case t := <-sched.tc:
			done <- <-sched.immediately(t)
		case <-sched.ctx.Done():
			done <- sched.ctx.Err()
		}
	}()
	return done
}

func (sched *Scheduler) Do(fn func(time.Time)) error {
	sched.Run(fn)
	err := sched.Start()
	if err == ErrAlreadyRunning {
		err = nil
	}
	return err
}

func Forever() {
	select {}
}
