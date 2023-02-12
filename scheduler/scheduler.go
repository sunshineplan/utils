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
	sync.Mutex
	ticker    *time.Ticker
	schedules []Time

	fn     func(time.Time)
	ctx    context.Context
	cancel context.CancelFunc
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

func (sched *Scheduler) At(schedules ...Time) *Scheduler {
	sched.Lock()
	defer sched.Unlock()
	sched.schedules = append(sched.schedules, schedules...)
	return sched
}

func (sched *Scheduler) Clear() {
	sched.Lock()
	defer sched.Unlock()
	sched.schedules = nil
}

func (sched *Scheduler) Run(fn func(time.Time)) *Scheduler {
	sched.Lock()
	defer sched.Unlock()
	sched.fn = fn
	return sched
}

func (sched *Scheduler) init(fn bool) error {
	sched.Lock()
	defer sched.Unlock()

	if fn && sched.fn == nil {
		return ErrNoFunction
	} else if len(sched.schedules) == 0 {
		return ErrNoSchedule
	}
	if sched.ctx == nil || sched.ctx.Err() != nil {
		sched.ticker = time.NewTicker(time.Second)
		sched.ctx, sched.cancel = context.WithCancel(context.Background())
		now := time.Now()
		for _, i := range sched.schedules {
			if i, ok := i.(*tickerSched); ok {
				i.start = now
			}
		}
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
				sched.Lock()
				for _, i := range sched.schedules {
					if i.IsMatched(t) {
						go fn(t)
						break
					}
				}
				sched.Unlock()
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
			sched.Lock()
			defer sched.Unlock()
			for _, i := range sched.schedules {
				if i.IsMatched(t) {
					go func() {
						sched.fn(t)
						done <- nil
					}()
					return
				}
			}
		case <-sched.ctx.Done():
			done <- sched.ctx.Err()
		}
	}()
	return done
}
