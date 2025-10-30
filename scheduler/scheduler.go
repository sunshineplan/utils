package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrNoFunction indicates that no function has been registered to run.
	ErrNoFunction = errors.New("scheduler function is not set")
	// ErrNoSchedule indicates that no schedule has been configured.
	ErrNoSchedule = errors.New("no schedule has been added")
	// ErrAlreadyRunning indicates that the scheduler is already active.
	ErrAlreadyRunning = errors.New("scheduler is already running")
)

// Scheduler defines a flexible time-based job runner.
// It supports multiple schedules, condition combinations, missed-event handling,
// and optional structured debug logging.
type Scheduler struct {
	mu sync.Mutex

	tc chan Event    // event channel from global ticker
	fn []func(Event) // functions to execute on trigger

	ignoreMissed atomic.Bool // whether to skip missed executions

	sched complexSched

	ctx    context.Context
	cancel context.CancelFunc

	debugLogger *slog.Logger
}

// NewScheduler creates a new, uninitialized Scheduler instance.
func NewScheduler() *Scheduler {
	return &Scheduler{tc: make(chan Event, 1)}
}

// WithDebug attaches a slog.Logger for debug output.
func (sched *Scheduler) WithDebug(logger *slog.Logger) *Scheduler {
	sched.debugLogger = logger
	return sched
}

// SetIgnoreMissed sets whether missed schedule times should be ignored.
// If true, the scheduler will skip backlogged runs caused by delays.
func (sched *Scheduler) SetIgnoreMissed(ignore bool) *Scheduler {
	sched.ignoreMissed.Store(ignore)
	return sched
}

// debug logs a debug message if a logger is configured.
func (sched *Scheduler) debug(msg string, args ...any) {
	if sched.debugLogger != nil {
		sched.debugLogger.Debug(msg, args...)
	}
}

// At sets the scheduler to trigger when *any* of the provided schedules match.
// This is a logical OR of all schedules.
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

// AtCondition sets the scheduler to trigger only when *all* schedules match.
// This is a logical AND of all schedules.
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

// String returns a human-readable representation of the configured schedule.
func (sched *Scheduler) String() string {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	if sched.sched == nil {
		return ""
	}
	return sched.sched.String()
}

// Clear removes all schedules and stops any running context.
func (sched *Scheduler) Clear() {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	sched.sched = nil
	if sched.ctx != nil && sched.ctx.Err() == nil {
		sched.Stop()
	}
}

// Run registers one or more functions to be executed when the schedule triggers.
// Functions are executed in separate goroutines.
func (sched *Scheduler) Run(fn ...func(Event)) *Scheduler {
	sched.mu.Lock()
	defer sched.mu.Unlock()
	for _, fn := range fn {
		if fn != nil {
			sched.fn = append(sched.fn, fn)
		}
	}
	return sched
}

// init prepares the scheduler for execution.
// It validates configuration, initializes contexts, and subscribes for the next event.
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
		next := sched.sched.Next(t)
		sched.debug("Scheduler Next Run Time", "Name", sched.sched, "Next", next)
		subscribe(next, sched.tc)
		return nil
	}
	return ErrAlreadyRunning
}

// start launches the main loop that listens for Event notifications
// and executes registered functions upon each trigger.
func (sched *Scheduler) start(once bool) error {
	if err := sched.init(); err != nil {
		return err
	}
	go func() {
		for {
			select {
			case e := <-sched.tc:
				if e.Missed {
					sched.debug("Scheduler Missed Run Time", "Name", sched.sched, "Time", e.Time, "Goal", e.Goal)
				}
				sched.mu.Lock()
				if once || !e.Missed || !sched.ignoreMissed.Load() {
					for _, fn := range sched.fn {
						go fn(e)
					}
				}
				sched.mu.Unlock()
				if once {
					unsubscribe(sched.tc)
					return
				}
				next := sched.sched.Next(e.Time)
				if next.IsZero() {
					sched.debug("Scheduler No More Next", "Name", sched.sched)
					unsubscribe(sched.tc)
					return
				}
				sched.debug("Scheduler Next Run Time", "Name", sched.sched, "Next", next)
				subscribe(next, sched.tc)
			case <-sched.ctx.Done():
				unsubscribe(sched.tc)
				return
			}
		}
	}()
	return nil
}

// Start begins running the scheduler continuously until Stop is called.
func (sched *Scheduler) Start() error {
	return sched.start(false)
}

// Once starts the scheduler for a single execution and returns an error channel
// that reports initialization status.
func (sched *Scheduler) Once() <-chan error {
	c := make(chan error, 1)
	go func() {
		c <- sched.start(true)
	}()
	return c
}

// Do runs the given function according to the current schedule configuration.
// If the scheduler is already running, the error is suppressed.
func (sched *Scheduler) Do(fn func(Event)) error {
	sched.Run(fn)
	err := sched.Start()
	if err == ErrAlreadyRunning {
		err = nil
	}
	return err
}

// Stop stops the scheduler and cancels its running context.
func (sched *Scheduler) Stop() {
	sched.cancel()
}

// immediately triggers all registered functions immediately with the given time,
// returning a channel that closes when all functions complete.
func (sched *Scheduler) immediately(t time.Time) <-chan struct{} {
	sched.mu.Lock()
	defer sched.mu.Unlock()

	var wg sync.WaitGroup
	wg.Add(len(sched.fn))
	for _, fn := range sched.fn {
		wg.Go(func() { fn(Event{Time: t, Goal: t}) })
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}

// Immediately triggers all registered functions right now (non-scheduled).
func (sched *Scheduler) Immediately() <-chan struct{} {
	return sched.immediately(time.Now())
}

// Forever blocks indefinitely, keeping the current goroutine alive.
func Forever() {
	select {}
}
