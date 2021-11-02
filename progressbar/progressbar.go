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
	"text/template"
	"time"

	"github.com/sunshineplan/utils/unit"
)

const defaultTemplate = `[{{.Done}}{{.Undone}}]   {{.Speed}}   {{.Current -}}
({{.Percent}}) of {{.Total}}   {{.Elapsed}}   {{.Left}} `

// ProgressBar is a simple progress bar.
type ProgressBar struct {
	sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	start      time.Time
	blockWidth int
	refresh    time.Duration
	template   *template.Template
	current    int64
	total      int64
	lastWidth  int
	speed      float64
	unit       string
}

type counter struct{ *ProgressBar }

var _ io.Writer = &counter{}

func (c *counter) Write(b []byte) (int, error) {
	c.Add(int64(len(b)))

	return 0, nil
}

type format struct {
	Done, Undone   string
	Speed, Percent string
	Current, Total string
	Elapsed, Left  string
}

// New returns a new ProgressBar with default options.
func New(total int) *ProgressBar {
	return New64(int64(total))
}

// New64 returns a new ProgressBar with default options.
func New64(total int64) *ProgressBar {
	if total <= 0 {
		panic("invalid total number")
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &ProgressBar{
		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}, 1),

		blockWidth: 40,
		refresh:    5 * time.Second,
		template:   template.Must(template.New("ProgressBar").Parse(defaultTemplate)),
		total:      total,
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
func (pb *ProgressBar) Add(num int64) {
	pb.Lock()
	defer pb.Unlock()

	pb.current += num
}

func (pb *ProgressBar) print(f format) {
	var buf bytes.Buffer
	pb.template.Execute(&buf, f)

	width := buf.Len()
	if width < pb.lastWidth {
		io.WriteString(os.Stderr,
			fmt.Sprintf("\r%s\r%s", strings.Repeat(" ", pb.lastWidth), buf.Bytes()))
	} else {
		io.WriteString(os.Stderr, "\r\r"+buf.String())
	}

	pb.lastWidth = width
}

func (pb *ProgressBar) startRefresh() {
	start := time.Now()
	maxRefresh := pb.refresh * 3

	ticker := time.NewTicker(pb.refresh)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pb.Lock()
			now := pb.current
			totalSpeed := float64(now) / (float64(time.Since(start)) / float64(time.Second))
			intervalSpeed := float64(pb.current-now) / (float64(pb.refresh) / float64(time.Second))
			if intervalSpeed == 0 {
				pb.speed = totalSpeed
			} else {
				pb.speed = intervalSpeed
			}
			if intervalSpeed == 0 && pb.refresh < maxRefresh {
				pb.refresh += time.Second
			}
			pb.Unlock()
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
			pb.Lock()

			now := pb.current
			if now > pb.total {
				now = pb.total
			}
			done := int(int64(pb.blockWidth) * now / pb.total)
			percent := float64(now) * 100 / float64(pb.total)

			var left time.Duration
			if pb.speed == 0 {
				left = 0
			} else {
				left = time.Duration(float64(pb.total-now)/pb.speed) * time.Second
			}

			var progressed string
			if now < pb.total && done != 0 {
				progressed = strings.Repeat("=", done-1) + ">"
			} else {
				progressed = strings.Repeat("=", done)
			}

			var f format
			if pb.unit == "bytes" {
				f = format{
					Done:    progressed,
					Undone:  strings.Repeat(" ", pb.blockWidth-done),
					Speed:   unit.FormatBytes(int64(pb.speed)) + "/s",
					Current: unit.FormatBytes(now),
					Percent: fmt.Sprintf("%.2f%%", percent),
					Total:   unit.FormatBytes(pb.total),
					Elapsed: fmt.Sprintf("Elapsed: %s", unit.FormatDuration(time.Since(pb.start))),
					Left:    fmt.Sprintf("Left: %s", unit.FormatDuration(left)),
				}
			} else {
				f = format{
					Done:    progressed,
					Undone:  strings.Repeat(" ", pb.blockWidth-done),
					Speed:   fmt.Sprintf("%.2f/s", pb.speed),
					Current: strconv.FormatInt(now, 10),
					Percent: fmt.Sprintf("%.2f%%", percent),
					Total:   strconv.FormatInt(pb.total, 10),
					Elapsed: fmt.Sprintf("Elapsed: %s", unit.FormatDuration(time.Since(pb.start))),
					Left:    fmt.Sprintf("Left: %s", unit.FormatDuration(left)),
				}
			}

			if pb.speed == 0 {
				f.Speed = "--/s"
				f.Left = "Left: calculating" + strings.Repeat(".", time.Now().Second()%3+1)
			}

			pb.print(f)

			if now == pb.total {
				totalSpeed := float64(pb.total) / (float64(time.Since(pb.start)) / float64(time.Second))
				if pb.unit == "bytes" {
					f.Speed = unit.FormatBytes(int64(totalSpeed)) + "/s"
				} else {
					f.Speed = fmt.Sprintf("%.2f/s", totalSpeed)
				}
				f.Left = "Complete"

				pb.print(f)
				io.WriteString(os.Stderr, "\n")

				pb.Unlock()
				close(pb.done)
				return
			}

			pb.Unlock()
		case <-pb.ctx.Done():
			io.WriteString(os.Stderr, "\nCancelled\n")
			return
		}
	}
}

// Start starts the progress bar.
func (pb *ProgressBar) Start() error {
	if !pb.start.IsZero() {
		return fmt.Errorf("progress bar is already started")
	}

	if pb.total < 0 {
		return fmt.Errorf("illegal total number: %d", pb.total)
	}

	pb.start = time.Now()

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

	return io.Copy(w, io.TeeReader(r, &counter{pb}))
}
