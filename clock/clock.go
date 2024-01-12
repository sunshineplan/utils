package clock

import (
	"cmp"
	"encoding"
	"fmt"
	"time"
)

const (
	secondsPerMinute = 60
	secondsPerHour   = 60 * secondsPerMinute
	secondsPerDay    = 24 * secondsPerHour
)

var (
	_ encoding.TextMarshaler   = Clock{}
	_ encoding.TextUnmarshaler = new(Clock)
)

type Clock struct {
	wall uint64
}

func abs[Int int | uint64](wall Int) uint64 {
	for wall < 0 {
		wall += secondsPerDay
	}
	return uint64(wall % secondsPerDay)
}

func New(hour, min, sec int) Clock {
	return Clock{abs(hour*secondsPerHour + min*secondsPerMinute + sec)}
}

var clockLayout = []string{
	"15:04",
	time.TimeOnly,
	time.Kitchen,
	"03:04PM",
	"03:04:05PM",
}

func Parse(v string) (Clock, error) {
	for _, layout := range clockLayout {
		if t, err := time.Parse(layout, v); err == nil {
			return ParseTime(t), nil
		}
	}
	return Clock{}, fmt.Errorf("cannot parse %q as clock", v)
}

func ParseTime(t time.Time) Clock {
	return New(t.Clock())
}

func Now() Clock {
	return ParseTime(time.Now())
}

func (c Clock) Time() time.Time {
	return time.Unix(int64(c.wall), 0).UTC()
}

func (c Clock) Clock() (hour, min, sec int) {
	sec = int(c.wall)
	hour = sec / secondsPerHour
	sec -= hour * secondsPerHour
	min = sec / secondsPerMinute
	sec -= min * secondsPerMinute
	return
}

func (c Clock) Hour() int {
	return int(c.wall%secondsPerDay) / secondsPerHour
}

func (c Clock) Minute() int {
	return int(c.wall%secondsPerHour) / secondsPerMinute
}

func (c Clock) Second() int {
	return int(c.wall % secondsPerMinute)
}

func (c Clock) String() string {
	return fmt.Sprintf("%d:%02d:%02d", c.Hour(), c.Minute(), c.Second())
}

func (c Clock) MarshalText() (text []byte, err error) {
	text = []byte(c.String())
	return
}

func (c *Clock) UnmarshalText(text []byte) error {
	clock, err := Parse(string(text))
	if err != nil {
		return err
	}
	*c = clock
	return nil
}

func (c Clock) After(u Clock) bool {
	return c.wall > u.wall
}

func (c Clock) Before(u Clock) bool {
	return c.wall < u.wall
}

func (c Clock) Equal(u Clock) bool {
	return c.wall == u.wall
}

func (c Clock) Compare(u Clock) int {
	return cmp.Compare(c.wall, u.wall)
}

func (c Clock) Add(d time.Duration) Clock {
	return Clock{abs(c.wall + uint64(d.Seconds()))}
}

func (c Clock) Sub(u Clock) time.Duration {
	return time.Duration(int64(c.wall)-int64(u.wall)) * time.Second
}

func (c Clock) Since(u Clock) time.Duration {
	return u.Until(c)
}

func (c Clock) Until(u Clock) time.Duration {
	d := int64(u.wall) - int64(c.wall)
	if d == 0 {
		return 0
	} else if d < 0 {
		d += secondsPerDay
	}
	return time.Duration(d) * time.Second
}

func Since(c Clock) time.Duration {
	return c.Until(Now())
}

func Until(c Clock) time.Duration {
	return Now().Until(c)
}
