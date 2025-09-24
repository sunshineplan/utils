package progressbar

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/sunshineplan/utils/counter"
	"github.com/sunshineplan/utils/unit"
)

const (
	defaultRefresh     = 5 * time.Second
	maxRefreshMultiple = 3
)

const defaultTemplate = `[{{.Done}}{{.Undone}}]  {{.Speed}}  {{.Current -}}
({{.Percent}}) of {{.Total}}{{if .Additional}} [{{.Additional}}]{{end}}  {{.Elapsed}}  {{.Left}} `

var dots = []string{".  ", ".. ", "..."}

// ProgressBar represents a customizable progress bar for tracking task progress.
// It supports configurable templates, units, and refresh intervals.
type ProgressBar struct {
	mu  sync.Mutex
	buf strings.Builder

	ctx       context.Context
	cancel    context.CancelFunc
	msgChan   chan string
	done      chan struct{}
	start     time.Time
	last      string
	lastWidth int

	blockWidth      int
	refreshInterval time.Duration
	renderInterval  time.Duration
	template        *template.Template

	current    counter.Counter
	total      int64
	additional string
	speed      float64
	unit       string
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
func New(total int) *ProgressBar {
	return New64(int64(total))
}

// New64 returns a new ProgressBar with default options.
func New64(total int64) *ProgressBar {
	if total <= 0 {
		panic(fmt.Sprintf("invalid total number: %d", total))
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &ProgressBar{
		ctx:             ctx,
		cancel:          cancel,
		msgChan:         make(chan string, 1),
		done:            make(chan struct{}),
		blockWidth:      40,
		refreshInterval: defaultRefresh,
		template:        template.Must(template.New("ProgressBar").Parse(defaultTemplate)),
		total:           int64(total),
	}
}

// SetWidth sets the progress bar block width.
// It panics if called after the progress bar has started or if blockWidth is less than or equal to zero.
func (pb *ProgressBar) SetWidth(blockWidth int) *ProgressBar {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	if !pb.start.IsZero() {
		panic("progress bar is already started")
	}
	if blockWidth <= 0 {
		panic(fmt.Sprintf("invalid block width: %d", blockWidth))
	}
	pb.blockWidth = blockWidth
	return pb
}

// SetRefreshInterval sets progress bar refresh interval time for check speed.
// It panics if called after the progress bar has started or if interval is less than or equal to zero.
func (pb *ProgressBar) SetRefreshInterval(interval time.Duration) *ProgressBar {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	if !pb.start.IsZero() {
		panic("progress bar is already started")
	}
	if interval <= 0 {
		panic(fmt.Sprintf("invalid refresh interval: %v", interval))
	}
	pb.refreshInterval = interval
	return pb
}

// SetRenderInterval sets the interval for updating the progress bar display.
// It panics if called after the progress bar has started or if interval is less than or equal to zero.
func (pb *ProgressBar) SetRenderInterval(interval time.Duration) *ProgressBar {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	if !pb.start.IsZero() {
		panic("progress bar is already started")
	}
	if interval <= 0 {
		panic(fmt.Sprintf("invalid render interval: %v", interval))
	}
	pb.renderInterval = interval
	return pb
}

// SetTemplate sets progress bar template.
func (pb *ProgressBar) SetTemplate(tmplt string) error {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	if !pb.start.IsZero() {
		return fmt.Errorf("progress bar is already started")
	}
	t := template.New("ProgressBar")
	if _, err := t.Parse(tmplt); err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	if err := t.Execute(io.Discard, format{}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	pb.template = t
	return nil
}

// SetUnit sets progress bar unit.
// It panics if called after the progress bar has started.
func (pb *ProgressBar) SetUnit(unit string) *ProgressBar {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	if !pb.start.IsZero() {
		panic("progress bar is already started")
	}
	pb.unit = unit
	return pb
}

// Add adds the specified amount to the progress bar.
func (pb *ProgressBar) Add(n int64) {
	pb.current.Add(n)
}

// Additional adds the specified string to the progress bar.
func (pb *ProgressBar) Additional(s string) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.additional = s
}

func (pb *ProgressBar) now() int64 {
	return pb.current.Load()
}

func (pb *ProgressBar) print(s string, msg bool) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
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

func (pb *ProgressBar) startRefresh() {
	start := pb.start
	maxRefresh := pb.refreshInterval * maxRefreshMultiple
	ticker := time.NewTicker(pb.refreshInterval)
	defer ticker.Stop()
	for {
		last := pb.now()
		select {
		case <-ticker.C:
			now := pb.now()
			totalSpeed := float64(now) / (float64(time.Since(start)) / float64(time.Second))
			intervalSpeed := float64(now-last) / (float64(pb.refreshInterval) / float64(time.Second))
			pb.mu.Lock()
			if intervalSpeed == 0 {
				pb.speed = totalSpeed
			} else {
				pb.speed = intervalSpeed
			}
			pb.mu.Unlock()
			if intervalSpeed == 0 && pb.refreshInterval < maxRefresh {
				pb.refreshInterval += time.Second
				ticker.Reset(pb.refreshInterval)
			}
		case <-pb.ctx.Done():
			return
		case <-pb.done:
			return
		}
	}
}

func (pb *ProgressBar) startCount() {
	interval := pb.renderInterval
	if interval == 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer func() {
		ticker.Stop()
		close(pb.done)
		close(pb.msgChan)
	}()
	var lastNow int64
	var f format
	if pb.unit == "bytes" {
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
			pb.mu.Lock()
			if now != lastNow || f.Done == "" {
				lastNow = now
				done := int(int64(pb.blockWidth) * now / pb.total)
				percent := float64(now) * 100 / float64(pb.total)
				if now < pb.total && done != 0 {
					f.Done = strings.Repeat("=", done-1) + ">"
				} else {
					f.Done = strings.Repeat("=", done)
				}
				f.Undone = strings.Repeat(" ", pb.blockWidth-done)
				f.Percent = fmt.Sprintf("%.2f%%", percent)
				if pb.unit == "bytes" {
					f.Current = unit.ByteSize(now).String()
				} else {
					f.Current = strconv.FormatInt(now, 10)
				}
			}
			f.Additional = pb.additional
			f.Elapsed = fmt.Sprintf("Elapsed: %s", time.Since(pb.start).Truncate(time.Second))
			if pb.speed == 0 {
				f.Speed = "--/s"
				f.Left = "Left: calculating" + dots[dot%3]
				dot++
			} else {
				if pb.unit == "bytes" {
					f.Speed = unit.ByteSize(pb.speed).String() + "/s"
				} else {
					f.Speed = fmt.Sprintf("%.2f/s", pb.speed)
				}
				f.Left = fmt.Sprintf("Left: %s", (time.Duration(float64(pb.total-now)/pb.speed) * time.Second).Truncate(time.Second))
			}
			pb.mu.Unlock()
			buf.Reset()
			pb.template.Execute(&buf, f)
			pb.print(buf.String(), false)
			if now == pb.total {
				totalSpeed := float64(pb.total) / (float64(time.Since(pb.start)) / float64(time.Second))
				if pb.unit == "bytes" {
					f.Speed = unit.ByteSize(totalSpeed).String() + "/s"
				} else {
					f.Speed = fmt.Sprintf("%.2f/s", totalSpeed)
				}
				f.Left = "Complete"
				buf.Reset()
				pb.template.Execute(&buf, f)
				pb.print(buf.String(), false)
				io.WriteString(os.Stdout, "\n")
				return
			}
		case msg := <-pb.msgChan:
			pb.print(msg, true)
		case <-pb.ctx.Done():
			pb.mu.Lock()
			defer pb.mu.Unlock()
			io.WriteString(os.Stdout, "\nCancelled\n")
			return
		}
	}
}

// Start starts the progress bar.
func (pb *ProgressBar) Start() error {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	if !pb.start.IsZero() {
		return fmt.Errorf("progress bar is already started")
	}
	pb.start = time.Now()
	go pb.startRefresh()
	go pb.startCount()
	return nil
}

// Message sets a message to be displayed on the progress bar.
func (pb *ProgressBar) Message(msg string) error {
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

// Done waits the progress bar finished. Same as [ProgressBar.Wait]().
func (pb *ProgressBar) Done() {
	pb.Wait()
}

// Wait blocks until the progress bar is finished.
func (pb *ProgressBar) Wait() {
	<-pb.done
}

// Cancel cancels the progress bar.
func (pb *ProgressBar) Cancel() {
	pb.cancel()
}

// FromReader starts the progress bar from a reader.
func (pb *ProgressBar) FromReader(r io.Reader, w io.Writer) (int64, error) {
	if err := pb.Start(); err != nil {
		return 0, err
	}
	n, err := io.Copy(pb.current.AddWriter(w), r)
	if err != nil {
		pb.Cancel()
		return n, err
	}
	return n, nil
}
