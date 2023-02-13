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
	mu sync.Mutex

	ticker *time.Ticker
	sched  complexSched

	fn     func(time.Time)
	ctx    context.Context
	cancel context.CancelFunc
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
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

func (sched *Scheduler) Run(fn func(time.Time)) *Scheduler {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.fn = fn
	return sched
}

func (sched *Scheduler) init(fn bool) error {
	sched.mu.Lock()
	defer sched.mu.Unlock()

	if fn && sched.fn == nil {
		return ErrNoFunction
	} else if sched.sched.len() == 0 {
		return ErrNoSchedule
	}
	if sched.ctx == nil || sched.ctx.Err() != nil {
		sched.ticker = time.NewTicker(time.Second)
		sched.ctx, sched.cancel = context.WithCancel(context.Background())
		sched.sched.init(time.Now())
		return nil
	}
	return ErrAlreadyRunning
}

func (sched *Scheduler) start(fn func(time.Time)) {
	if fn == nil {
		panic("function cannot be nil")
	}
	go func() {
		for {
			select {
			case t := <-sched.ticker.C:
				sched.mu.Lock()
				if sched.sched.IsMatched(t) {
					go fn(t)
				}
				sched.mu.Unlock()
			case <-sched.ctx.Done():
				return
			}
		}
	}()
}

func (sched *Scheduler) Start() error {
	if err := sched.init(true); err != nil {
		return err
	}
	sched.start(sched.fn)
	return nil
}

func (sched *Scheduler) Stop() {
	sched.ticker.Stop()
	sched.cancel()
}

func (sched *Scheduler) Do(fn func(time.Time)) error {
	if err := sched.init(false); err == ErrNoSchedule {
		return err
	}
	sched.start(fn)
	return nil
}

func (sched *Scheduler) Immediately() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		sched.fn(time.Now())
		done <- struct{}{}
	}()
	return done
}

func (sched *Scheduler) Once() <-chan error {
	done := make(chan error)
	if err := sched.init(true); err != nil {
		done <- err
		return done
	}
	go func() {
		select {
		case t := <-sched.ticker.C:
			sched.mu.Lock()
			defer sched.mu.Unlock()
			if sched.sched.IsMatched(t) {
				go func() {
					sched.fn(t)
					done <- nil
				}()
				return
			}
		case <-sched.ctx.Done():
			done <- sched.ctx.Err()
		}
	}()
	return done
}
