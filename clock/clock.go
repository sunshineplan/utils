package clock

import (
	"cmp"
	"encoding"
	"fmt"
	"time"
	"unique"
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

var w0 unique.Handle[uint64]

type Clock struct {
	wall unique.Handle[uint64]
}

func wall(i int64) unique.Handle[uint64] {
	if i %= secondsPerDay; i < 0 {
		i += secondsPerDay
	}
	return unique.Make(uint64(i))
}

func New(hour, min, sec int) Clock {
	return Clock{wall(int64(hour)*secondsPerHour + int64(min)*secondsPerMinute + int64(sec))}
}

var clockLayout = []string{
	"3:04PM", // [time.Kitchen]
	"3:04:05PM",
	"15:04",
	"15:04:05", // [time.TimeOnly]
}

func Parse(v string) (Clock, error) {
	for _, layout := range clockLayout {
		if t, err := time.Parse(layout, v); err == nil {
			return ParseTime(t), nil
		}
	}
	return Clock{}, fmt.Errorf("cannot parse %q as clock", v)
}

func MustParse(v string) Clock {
	if c, err := Parse(v); err != nil {
		panic("clock: " + err.Error())
	} else {
		return c
	}
}

func ParseTime(t time.Time) Clock {
	return New(t.Clock())
}

func Now() Clock {
	return ParseTime(time.Now())
}

func (c Clock) Time() time.Time {
	return time.Unix(int64(c.wall.Value()), 0).UTC()
}

func (c Clock) Clock() (hour, min, sec int) {
	sec = int(c.wall.Value())
	hour = sec / secondsPerHour
	sec -= hour * secondsPerHour
	min = sec / secondsPerMinute
	sec -= min * secondsPerMinute
	return
}

func (c Clock) Seconds() uint64 {
	return c.wall.Value()
}

func (c Clock) Hour() int {
	return int(c.Seconds()%secondsPerDay) / secondsPerHour
}

func (c Clock) Minute() int {
	return int(c.Seconds()%secondsPerHour) / secondsPerMinute
}

func (c Clock) Second() int {
	return int(c.Seconds() % secondsPerMinute)
}

func (c Clock) String() string {
	if c.wall == w0 {
		return "invalid"
	}
	return fmt.Sprintf("%d:%02d:%02d", c.Hour(), c.Minute(), c.Second())
}

func (c Clock) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

func (c *Clock) UnmarshalText(text []byte) error {
	clock, err := Parse(string(text))
	if err != nil {
		return err
	}
	*c = clock
	return nil
}

func (c Clock) IsValid() bool {
	return c.wall != w0
}

func (c Clock) After(u Clock) bool {
	return c.wall.Value() > u.wall.Value()
}

func (c Clock) Before(u Clock) bool {
	return c.wall.Value() < u.wall.Value()
}

func (c Clock) Equal(u Clock) bool {
	return c.wall == u.wall
}

func (c Clock) Compare(u Clock) int {
	return cmp.Compare(c.wall.Value(), u.wall.Value())
}

func (c Clock) Add(d time.Duration) Clock {
	return Clock{wall(int64(c.wall.Value()) + int64(d.Seconds()))}
}

func (c Clock) Sub(u Clock) time.Duration {
	return time.Duration(int64(c.Seconds())-int64(u.Seconds())) * time.Second
}

func (c Clock) Since(u Clock) time.Duration {
	return u.Until(c)
}

func (c Clock) Until(u Clock) time.Duration {
	d := int64(u.Seconds()) - int64(c.Seconds())
	if d < 0 {
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
