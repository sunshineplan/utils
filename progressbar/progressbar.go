package progressbar

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/sunshineplan/utils/container"
	"github.com/sunshineplan/utils/unit"
)

const (
	defaultBlockWidth  = 40
	defaultRefresh     = 5 * time.Second
	defaultRender      = time.Second
	maxRefreshMultiple = 3
)

const defaultTemplate = `[{{.Done}}{{.Undone}}]  {{.Speed}}  {{.Current -}}
({{.Percent}}) of {{.Total}}{{if .Additional}} [{{.Additional}}]{{end}}  {{.Elapsed}}  {{.Left}} `

var dots = []string{".  ", ".. ", "..."}

// ProgressBar represents a customizable progress bar for tracking task progress.
// It supports configurable templates, units, and refresh intervals.
type ProgressBar[T int | int64] struct {
	buf strings.Builder

	ctx       context.Context
	cancel    context.CancelFunc
	msgChan   chan string
	done      chan struct{}
	start     container.Value[time.Time]
	last      string
	lastWidth int

	blockWidth      container.Int[int]
	refreshInterval container.Int[time.Duration]
	renderInterval  container.Int[time.Duration]
	template        atomic.Pointer[template.Template]
	unit            container.Value[string]

	current    genericCounter
	total      int64
	additional container.Value[string]
	speed      container.Value[float64]
}

type format struct {
	Done, Undone   string
	Speed, Percent string
	Current, Total string
	Additional     string
	Elapsed, Left  string
}

// New creates a new ProgressBar with the specified total count and default options.
// It panics if total is less than or equal to zero.
func New[T int | int64](total T) *ProgressBar[T] {
	if total <= 0 {
		panic(fmt.Sprintf("invalid total number: %d", total))
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &ProgressBar[T]{
		ctx:     ctx,
		cancel:  cancel,
		msgChan: make(chan string, 1),
		done:    make(chan struct{}),
		current: newNumberCounter(),
		total:   int64(total),
	}
}

// SetWidth sets the progress bar block width.
// It panics if blockWidth is less than or equal to zero.
func (pb *ProgressBar[T]) SetWidth(blockWidth int) *ProgressBar[T] {
	if blockWidth <= 0 {
		panic(fmt.Sprintf("invalid block width: %d", blockWidth))
	}
	pb.blockWidth.Store(blockWidth)
	return pb
}

// SetRefreshInterval sets progress bar refresh interval time for check speed.
// It panics if called after the progress bar has started or if interval is less than or equal to zero.
func (pb *ProgressBar[T]) SetRefreshInterval(interval time.Duration) *ProgressBar[T] {
	if !pb.start.Load().IsZero() {
		panic("progress bar is already started")
	}
	if interval <= 0 {
		panic(fmt.Sprintf("invalid refresh interval: %s", interval))
	}
	pb.refreshInterval.Store(interval)
	return pb
}

// SetRenderInterval sets the interval for updating the progress bar display.
// It panics if called after the progress bar has started or if interval is less than or equal to zero.
func (pb *ProgressBar[T]) SetRenderInterval(interval time.Duration) *ProgressBar[T] {
	if !pb.start.Load().IsZero() {
		panic("progress bar is already started")
	}
	if interval <= 0 {
		panic(fmt.Sprintf("invalid render interval: %s", interval))
	}
	pb.renderInterval.Store(interval)
	return pb
}

// SetTemplate sets progress bar template.
func (pb *ProgressBar[T]) SetTemplate(tmplt string) error {
	t := template.New("ProgressBar")
	if _, err := t.Parse(tmplt); err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	if err := t.Execute(io.Discard, format{}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	pb.template.Store(t)
	return nil
}

// SetUnit sets progress bar unit.
func (pb *ProgressBar[T]) SetUnit(unit string) *ProgressBar[T] {
	pb.unit.Store(unit)
	return pb
}

// Add adds the specified amount to the progress bar.
func (pb *ProgressBar[T]) Add(n T) {
	pb.current.Add(int64(n))
}

// Additional adds the specified string to the progress bar.
func (pb *ProgressBar[T]) Additional(s string) {
	pb.additional.Store(s)
}

func (pb *ProgressBar[T]) now() int64 {
	return pb.current.Get()
}

func (pb *ProgressBar[T]) print(s string, msg bool) {
	pb.buf.Reset()
	if len(s) < pb.lastWidth {
		pb.buf.WriteRune('\r')
		pb.buf.WriteString(strings.Repeat(" ", pb.lastWidth))
		pb.buf.WriteRune('\r')
		pb.buf.WriteString(s)
	} else {
		pb.buf.WriteRune('\r')
		pb.buf.WriteString(s)
	}
	if msg {
		pb.buf.WriteRune('\n')
		pb.buf.WriteString(pb.last)
	} else {
		pb.last = s
		pb.lastWidth = len(s)
	}
	io.WriteString(os.Stdout, pb.buf.String())
}

func (pb *ProgressBar[T]) startRefresh() {
	start := pb.start.Load()
	maxRefresh := pb.refreshInterval.Load() * maxRefreshMultiple
	ticker := time.NewTicker(pb.refreshInterval.Load())
	defer ticker.Stop()
	for {
		last := pb.now()
		select {
		case <-ticker.C:
			now := pb.now()
			totalSpeed := float64(now) / (float64(time.Since(start)) / float64(time.Second))
			intervalSpeed := float64(now-last) / (float64(pb.refreshInterval.Load()) / float64(time.Second))
			if intervalSpeed == 0 {
				pb.speed.Store(totalSpeed)
			} else {
				pb.speed.Store(intervalSpeed)
			}
			if intervalSpeed == 0 && pb.refreshInterval.Load() < maxRefresh {
				d := pb.refreshInterval.Add(time.Second)
				ticker.Reset(d)
			}
		case <-pb.ctx.Done():
			return
		case <-pb.done:
			return
		}
	}
}

func (pb *ProgressBar[T]) startCount() {
	ticker := time.NewTicker(pb.renderInterval.Load())
	defer func() {
		ticker.Stop()
		close(pb.done)
		close(pb.msgChan)
	}()
	var lastNow int64
	var f format
	if pb.unit.Load() == "bytes" {
		f.Total = unit.ByteSize(pb.total).String()
	} else {
		f.Total = strconv.FormatInt(pb.total, 10)
	}
	var buf strings.Builder
	var dot int
	for {
		select {
		case <-ticker.C:
			now := min(pb.now(), pb.total)
			if now != lastNow || f.Done == "" {
				lastNow = now
				blockWidth := pb.blockWidth.Load()
				done := int(int64(blockWidth) * now / pb.total)
				percent := float64(now) * 100 / float64(pb.total)
				if now < pb.total && done != 0 {
					f.Done = strings.Repeat("=", done-1) + ">"
				} else {
					f.Done = strings.Repeat("=", done)
				}
				f.Undone = strings.Repeat(" ", blockWidth-done)
				f.Percent = fmt.Sprintf("%.2f%%", percent)
				if pb.unit.Load() == "bytes" {
					f.Current = unit.ByteSize(now).String()
				} else {
					f.Current = strconv.FormatInt(now, 10)
				}
			}
			f.Additional = pb.additional.Load()
			f.Elapsed = fmt.Sprintf("Elapsed: %s", time.Since(pb.start.Load()).Truncate(time.Second))
			if speed := pb.speed.Load(); speed == 0 {
				f.Speed = "--/s"
				f.Left = "Left: calculating" + dots[dot%3]
				dot++
			} else {
				if pb.unit.Load() == "bytes" {
					f.Speed = unit.ByteSize(speed).String() + "/s"
				} else {
					f.Speed = fmt.Sprintf("%.2f/s", speed)
				}
				f.Left = fmt.Sprintf("Left: %s", (time.Duration(float64(pb.total-now)/speed) * time.Second).Truncate(time.Second))
			}
			buf.Reset()
			pb.template.Load().Execute(&buf, f)
			pb.print(buf.String(), false)
			if now == pb.total {
				totalSpeed := float64(pb.total) / (float64(time.Since(pb.start.Load())) / float64(time.Second))
				if pb.unit.Load() == "bytes" {
					f.Speed = unit.ByteSize(totalSpeed).String() + "/s"
				} else {
					f.Speed = fmt.Sprintf("%.2f/s", totalSpeed)
				}
				f.Left = "Complete"
				buf.Reset()
				pb.template.Load().Execute(&buf, f)
				pb.print(buf.String(), false)
				io.WriteString(os.Stdout, "\n")
				return
			}
		case msg := <-pb.msgChan:
			pb.print(msg, true)
		case <-pb.ctx.Done():
			io.WriteString(os.Stdout, "\nCancelled\n")
			return
		}
	}
}

// Start starts the progress bar.
func (pb *ProgressBar[T]) Start() error {
	if !pb.start.Load().IsZero() {
		return fmt.Errorf("progress bar is already started")
	}
	if pb.blockWidth.Load() == 0 {
		pb.blockWidth.Store(defaultBlockWidth)
	}
	if pb.renderInterval.Load() == 0 {
		pb.renderInterval.Store(defaultRender)
	}
	if pb.refreshInterval.Load() == 0 {
		pb.refreshInterval.Store(defaultRefresh)
	}
	if pb.template.Load() == nil {
		pb.template.Store(template.Must(template.New("ProgressBar").Parse(defaultTemplate)))
	}
	pb.start.Store(time.Now())
	go pb.startRefresh()
	go pb.startCount()
	return nil
}

// Message sets a message to be displayed on the progress bar.
func (pb *ProgressBar[T]) Message(msg string) error {
	defer func() { recover() }()
	select {
	case <-pb.done:
		return fmt.Errorf("progress bar is already finished")
	default:
	}
	select {
	case pb.msgChan <- msg:
	default:
	}
	return nil
}

// Wait blocks until the progress bar is finished.
func (pb *ProgressBar[T]) Wait() {
	<-pb.done
}

// Cancel cancels the progress bar.
func (pb *ProgressBar[T]) Cancel() {
	pb.cancel()
}

// FromReader starts the progress bar from a reader.
func (pb *ProgressBar[T]) FromReader(r io.Reader, w io.Writer) (int64, error) {
	pb.current = newWriterCounter(w)
	if err := pb.Start(); err != nil {
		return 0, err
	}
	n, err := io.Copy(pb.current, r)
	if err != nil {
		pb.Cancel()
	}
	return n, err
}
