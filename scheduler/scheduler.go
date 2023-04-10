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

	ctx    context.Context
	cancel context.CancelFunc
}

func NewScheduler() *Scheduler {
	return &Scheduler{notify: make(chan time.Time, 1), tc: make(chan time.Time, 1)}
}

func (sched *Scheduler) At(schedules ...Schedule) *Scheduler {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.sched = multiSched(schedules)
	return sched
}

func (sched *Scheduler) AtCondition(schedules ...Schedule) *Scheduler {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.sched = condSched(schedules)
	return sched
}

func (sched *Scheduler) Clear() {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.sched = nil
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

func (sched *Scheduler) init(d time.Duration, t time.Time) error {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	if len(sched.fn) == 0 {
		return ErrNoFunction
	} else if sched.sched.len() == 0 {
		return ErrNoSchedule
	}
	if sched.ctx == nil || sched.ctx.Err() != nil {
		sched.ctx, sched.cancel = context.WithCancel(context.Background())
		sched.sched.init(t)
		subscribeNotify(sched.notify)
		sched.newTimer(sched.sched.First(t), d)
		return nil
	}
	return ErrAlreadyRunning
}

func (sched *Scheduler) newTimer(first, duration time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	sched.timer = time.AfterFunc(first, func() {
		cancel()
		var now time.Time
		sched.ticker, now = time.NewTicker(duration), time.Now()
		go func() {
			for {
				select {
				case t := <-sched.ticker.C:
					sched.mu.Lock()
					if sched.sched.IsMatched(t) {
						sched.tc <- t
					}
					sched.mu.Unlock()
				case t := <-sched.notify:
					sched.mu.Lock()
					defer sched.mu.Unlock()
					sched.ticker.Stop()
					sched.newTimer(sched.sched.First(t), duration)
					return
				case <-sched.ctx.Done():
					sched.ticker.Stop()
					return
				}
			}
		}()
		sched.mu.Lock()
		defer sched.mu.Unlock()
		if sched.sched.IsMatched(now) {
			sched.tc <- now
		}
	})
	go func() {
		select {
		case t := <-sched.notify:
			if sched.timer.Stop() {
				sched.timer.Reset(sched.sched.First(t))
			}
		case <-sched.ctx.Done():
			cancel()
			sched.timer.Stop()
		case <-ctx.Done():
		}
	}()
}

func (sched *Scheduler) Start() error {
	if err := sched.init(sched.sched.TickerDuration(), time.Now()); err != nil {
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
		if err := sched.init(sched.sched.TickerDuration(), time.Now()); err != nil {
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
