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

// Clock represents a time within a day (hour, minute, second)
type Clock struct {
	wall unique.Handle[uint64]
}

func wall(i int64) unique.Handle[uint64] {
	if i %= secondsPerDay; i < 0 {
		i += secondsPerDay
	}
	return unique.Make(uint64(i))
}

// New constructs a Clock from hour, minute, second.
// Values can exceed normal ranges; they are normalized modulo 24 hours.
func New(hour, min, sec int) Clock {
	return Clock{wall(int64(hour)*secondsPerHour + int64(min)*secondsPerMinute + int64(sec))}
}

var clockLayout = []string{
	"3:04PM", // [time.Kitchen]
	"3:04:05PM",
	"15:04",
	"15:04:05", // [time.TimeOnly]
}

// Parse parses a string into a Clock using supported layouts
func Parse(v string) (Clock, error) {
	for _, layout := range clockLayout {
		if t, err := time.Parse(layout, v); err == nil {
			return ParseTime(t), nil
		}
	}
	return Clock{}, fmt.Errorf("cannot parse %q as clock", v)
}

// MustParse parses a string and panics if parsing fails
func MustParse(v string) Clock {
	if c, err := Parse(v); err != nil {
		panic("clock: " + err.Error())
	} else {
		return c
	}
}

// ParseTime converts a time.Time into a Clock (hour, minute, second)
func ParseTime(t time.Time) Clock {
	return New(t.Clock())
}

// Now returns the current local time as a Clock
func Now() Clock {
	return ParseTime(time.Now())
}

// Time converts the Clock into a time.Time on the Unix epoch date
func (c Clock) Time() time.Time {
	return time.Unix(int64(c.wall.Value()), 0).UTC()
}

// Clock returns hour, minute, second components of the Clock
func (c Clock) Clock() (hour, min, sec int) {
	sec = int(c.wall.Value())
	hour = sec / secondsPerHour
	sec -= hour * secondsPerHour
	min = sec / secondsPerMinute
	sec -= min * secondsPerMinute
	return
}

// Seconds returns the number of seconds
func (c Clock) Seconds() uint64 {
	return c.wall.Value()
}

// Hour returns the hour of the Clock
func (c Clock) Hour() int {
	return int(c.Seconds()%secondsPerDay) / secondsPerHour
}

// Minute returns the minute of the Clock
func (c Clock) Minute() int {
	return int(c.Seconds()%secondsPerHour) / secondsPerMinute
}

// Second returns the second of the Clock
func (c Clock) Second() int {
	return int(c.Seconds() % secondsPerMinute)
}

// String returns the Clock as a string "H:MM:SS", or "invalid" if zero
func (c Clock) String() string {
	if c.wall == w0 {
		return "invalid"
	}
	return fmt.Sprintf("%d:%02d:%02d", c.Hour(), c.Minute(), c.Second())
}

// MarshalText implements encoding.TextMarshaler
func (c Clock) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (c *Clock) UnmarshalText(text []byte) error {
	clock, err := Parse(string(text))
	if err != nil {
		return err
	}
	*c = clock
	return nil
}

// IsValid returns true if the Clock is not the zero/invalid value
func (c Clock) IsValid() bool {
	return c.wall != w0
}

// After returns true if c is after u
func (c Clock) After(u Clock) bool {
	return c.wall.Value() > u.wall.Value()
}

// Before returns true if c is before u
func (c Clock) Before(u Clock) bool {
	return c.wall.Value() < u.wall.Value()
}

// Equal returns true if c equals u
func (c Clock) Equal(u Clock) bool {
	return c.wall == u.wall
}

// Compare compares c with u and returns -1,0,1
func (c Clock) Compare(u Clock) int {
	return cmp.Compare(c.wall.Value(), u.wall.Value())
}

// Add adds a duration to the Clock and returns a new Clock
// Duration is truncated to seconds
func (c Clock) Add(d time.Duration) Clock {
	return Clock{wall(int64(c.wall.Value()) + int64(d.Seconds()))}
}

// Sub returns the duration between c and u (c - u)
func (c Clock) Sub(u Clock) time.Duration {
	return time.Duration(int64(c.Seconds())-int64(u.Seconds())) * time.Second
}

// Since returns the duration from u until c (c - u, wrapped around 24h)
func (c Clock) Since(u Clock) time.Duration {
	return u.Until(c)
}

// Until returns the duration from c until u (always non-negative, wraps around 24h)
func (c Clock) Until(u Clock) time.Duration {
	d := int64(u.Seconds()) - int64(c.Seconds())
	if d < 0 {
		d += secondsPerDay
	}
	return time.Duration(d) * time.Second
}

// Since returns duration from u until now
func Since(c Clock) time.Duration {
	return c.Until(Now())
}

// Until returns duration from now until c
func Until(c Clock) time.Duration {
	return Now().Until(c)
}
