package progressbar

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/sunshineplan/utils/counter"
	"github.com/sunshineplan/utils/unit"
)

const defaultTemplate = `[{{.Done}}{{.Undone}}]  {{.Speed}}  {{.Current -}}
({{.Percent}}) of {{.Total}}{{if .Additional}} [{{.Additional}}]{{end}}  {{.Elapsed}}  {{.Left}} `

// ProgressBar is a simple progress bar.
type ProgressBar struct {
	mu sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	start      time.Time
	blockWidth int
	refresh    time.Duration
	template   *template.Template
	current    counter.Counter
	total      int64
	additional atomic.Value
	lastWidth  int
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

// New returns a new ProgressBar with default options.
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
		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}, 1),

		blockWidth: 40,
		refresh:    5 * time.Second,
		template:   template.Must(template.New("ProgressBar").Parse(defaultTemplate)),
		total:      int64(total),
	}
}

// SetWidth sets progress bar block width.
func (pb *ProgressBar) SetWidth(blockWidth int) *ProgressBar {
	pb.blockWidth = blockWidth

	return pb
}

// SetRefresh sets progress bar refresh time for check speed.
func (pb *ProgressBar) SetRefresh(refresh time.Duration) *ProgressBar {
	pb.refresh = refresh

	return pb
}

// SetTemplate sets progress bar template.
func (pb *ProgressBar) SetTemplate(tmplt string) (err error) {
	t := template.New("ProgressBar")
	if _, err = t.Parse(tmplt); err != nil {
		return
	}

	if err = t.Execute(io.Discard, format{}); err != nil {
		return
	}

	pb.template = t

	return
}

// SetUnit sets progress bar unit.
func (pb *ProgressBar) SetUnit(unit string) *ProgressBar {
	pb.unit = unit

	return pb
}

// Add adds the specified amount to the progress bar.
func (pb *ProgressBar) Add(n int64) {
	pb.current.Add(n)
}

// Additional adds the specified string to the progress bar.
func (pb *ProgressBar) Additional(s string) {
	pb.additional.Store(s)
}

func (pb *ProgressBar) now() int64 {
	return pb.current.Load()
}

func (pb *ProgressBar) print(f format) {
	var buf bytes.Buffer
	pb.template.Execute(&buf, f)

	width := buf.Len()
	if width < pb.lastWidth {
		io.WriteString(os.Stdout,
			fmt.Sprintf("\r%s\r%s", strings.Repeat(" ", pb.lastWidth), buf.Bytes()))
	} else {
		io.WriteString(os.Stdout, "\r\r"+buf.String())
	}

	pb.lastWidth = width
}

func (pb *ProgressBar) startRefresh() {
	start := time.Now()
	maxRefresh := pb.refresh * 3

	ticker := time.NewTicker(pb.refresh)
	defer ticker.Stop()

	for {
		last := pb.now()
		select {
		case <-ticker.C:
			now := pb.now()
			totalSpeed := float64(now) / (float64(time.Since(start)) / float64(time.Second))
			intervalSpeed := float64(now-last) / (float64(pb.refresh) / float64(time.Second))
			pb.mu.Lock()
			if intervalSpeed == 0 {
				pb.speed = totalSpeed
			} else {
				pb.speed = intervalSpeed
			}
			pb.mu.Unlock()
			if intervalSpeed == 0 && pb.refresh < maxRefresh {
				pb.refresh += time.Second
				ticker.Reset(pb.refresh)
			}
		case <-pb.ctx.Done():
			return
		case <-pb.done:
			return
		}
	}
}

func (pb *ProgressBar) startCount() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := min(pb.now(), pb.total)
			done := int(int64(pb.blockWidth) * now / pb.total)
			percent := float64(now) * 100 / float64(pb.total)

			var progressed string
			if now < pb.total && done != 0 {
				progressed = strings.Repeat("=", done-1) + ">"
			} else {
				progressed = strings.Repeat("=", done)
			}

			pb.mu.Lock()
			var left time.Duration
			if pb.speed != 0 {
				left = time.Duration(float64(pb.total-now)/pb.speed) * time.Second
			}

			var f format
			if pb.unit == "bytes" {
				f = format{
					Done:       progressed,
					Undone:     strings.Repeat(" ", pb.blockWidth-done),
					Speed:      unit.ByteSize(pb.speed).String() + "/s",
					Current:    unit.ByteSize(now).String(),
					Percent:    fmt.Sprintf("%.2f%%", percent),
					Total:      unit.ByteSize(pb.total).String(),
					Additional: pb.additional.Load().(string),
					Elapsed:    fmt.Sprintf("Elapsed: %s", time.Since(pb.start).Truncate(time.Second)),
					Left:       fmt.Sprintf("Left: %s", left.Truncate(time.Second)),
				}
			} else {
				f = format{
					Done:       progressed,
					Undone:     strings.Repeat(" ", pb.blockWidth-done),
					Speed:      fmt.Sprintf("%.2f/s", pb.speed),
					Current:    strconv.FormatInt(now, 10),
					Percent:    fmt.Sprintf("%.2f%%", percent),
					Total:      strconv.FormatInt(pb.total, 10),
					Additional: pb.additional.Load().(string),
					Elapsed:    fmt.Sprintf("Elapsed: %s", time.Since(pb.start).Truncate(time.Second)),
					Left:       fmt.Sprintf("Left: %s", left.Truncate(time.Second)),
				}
			}

			if pb.speed == 0 {
				f.Speed = "--/s"
				f.Left = fmt.Sprintf("Left: calculating%s%s",
					strings.Repeat(".", time.Now().Second()%3+1),
					strings.Repeat(" ", 2-time.Now().Second()%3),
				)
			}
			pb.mu.Unlock()

			pb.print(f)

			if now == pb.total {
				totalSpeed := float64(pb.total) / (float64(time.Since(pb.start)) / float64(time.Second))
				if pb.unit == "bytes" {
					f.Speed = unit.ByteSize(totalSpeed).String() + "/s"
				} else {
					f.Speed = fmt.Sprintf("%.2f/s", totalSpeed)
				}
				f.Left = "Complete"

				pb.print(f)
				io.WriteString(os.Stdout, "\n")

				close(pb.done)
				return
			}
		case <-pb.ctx.Done():
			io.WriteString(os.Stdout, "\nCancelled\n")
			return
		}
	}
}

// Start starts the progress bar.
func (pb *ProgressBar) Start() error {
	if !pb.start.IsZero() {
		return fmt.Errorf("progress bar is already started")
	}

	pb.start = time.Now()
	pb.additional.Store("")

	go pb.startRefresh()
	go pb.startCount()

	return nil
}

// Done waits the progress bar finished.
func (pb *ProgressBar) Done() {
	<-pb.done
}

// Cancel cancels the progress bar.
func (pb *ProgressBar) Cancel() {
	pb.cancel()
	close(pb.done)
}

// FromReader starts the progress bar from a reader.
func (pb *ProgressBar) FromReader(r io.Reader, w io.Writer) (int64, error) {
	pb.Start()
	return io.Copy(pb.current.AddWriter(w), r)
}
