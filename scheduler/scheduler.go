package scheduler

import (
	"context"
	"errors"
	"log/slog"
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

	tc chan time.Time
	fn []func(time.Time)

	sched complexSched
	next  time.Time

	ctx    context.Context
	cancel context.CancelFunc

	debugLogger *slog.Logger
}

func NewScheduler() *Scheduler {
	return &Scheduler{tc: make(chan time.Time, 1)}
}

func (sched *Scheduler) WithDebug(logger *slog.Logger) *Scheduler {
	sched.debugLogger = logger
	return sched
}

func (sched *Scheduler) debug(msg string, args ...any) {
	if sched.debugLogger != nil {
		sched.debugLogger.Debug(msg, args...)
	}
}

func (sched *Scheduler) At(schedules ...Schedule) *Scheduler {
	if len(schedules) == 0 {
		panic("no schedules")
	}
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.sched = multiSched(schedules)
	sched.debug("Scheduler At", "schedules", sched.sched)
	return sched
}

func (sched *Scheduler) AtCondition(schedules ...Schedule) *Scheduler {
	if len(schedules) == 0 {
		panic("no schedules")
	}
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.sched = condSched(schedules)
	sched.debug("Scheduler At Condition", "schedules", sched.sched)
	return sched
}

func (sched *Scheduler) String() string {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	if sched.sched == nil {
		return ""
	}
	return sched.sched.String()
}

func (sched *Scheduler) Clear() {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.sched = nil
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
		t := time.Now()
		sched.sched.init(t)
		sched.next = sched.sched.Next(t)
		sched.debug("Scheduler Next Run Time", "Name", sched.sched, "Next", sched.next)
		subscribe(sched.next, sched.tc)
		return nil
	}
	return ErrAlreadyRunning
}

func (sched *Scheduler) start(once bool) error {
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
				if once {
					unsubscribe(sched.tc)
					return
				}
				sched.next = sched.sched.Next(t)
				sched.debug("Scheduler Next Run Time", "Name", sched.sched, "Next", sched.next)
				subscribe(sched.next, sched.tc)
			case <-sched.ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (sched *Scheduler) Start() error {
	return sched.start(false)
}

func (sched *Scheduler) Once() <-chan error {
	c := make(chan error, 1)
	go func() {
		c <- sched.start(true)
	}()
	return c
}

func (sched *Scheduler) Do(fn func(time.Time)) error {
	sched.Run(fn)
	err := sched.Start()
	if err == ErrAlreadyRunning {
		err = nil
	}
	return err
}

func (sched *Scheduler) Stop() {
	sched.cancel()
	unsubscribe(sched.tc)
}

func (sched *Scheduler) immediately(t time.Time) <-chan struct{} {
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
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}

func (sched *Scheduler) Immediately() <-chan struct{} {
	return sched.immediately(time.Now())
}

func Forever() {
	select {}
}
