package clock

import (
	"testing"
	"time"
)

func TestClock(t *testing.T) {
	for i, tc := range []struct {
		hour, min, sec int
		c              Clock
		str            string
	}{
		{0, 0, 0, Clock{}, "0:00:00"},
		{1, 2, 3, New(1, 2, 3), "1:02:03"},
		{24, 0, 0, Clock{}, "0:00:00"},
		{0, 60, 0, New(1, 0, 0), "1:00:00"},
		{0, 0, 60, New(0, 1, 0), "0:01:00"},
		{0, 0, -1, New(23, 59, 59), "23:59:59"},
	} {
		if got := New(tc.hour, tc.min, tc.sec); got != tc.c {
			t.Errorf("#%d: New(%d, %d, %d): got %v; want %v", i, tc.hour, tc.min, tc.sec, got, tc.c)
		} else if got.String() != tc.str {
			t.Errorf("#%d: New(%d, %d, %d): got %q; want %q", i, tc.hour, tc.min, tc.sec, got.String(), tc.str)
		}
	}
}

func TestParse(t *testing.T) {
	for i, testcase := range []struct {
		s        string
		expected Clock
		str      string
	}{
		{"7:01", New(7, 1, 0), "7:01:00"},
		{"7:01:02", New(7, 1, 2), "7:01:02"},
		{"7:02PM", New(19, 2, 0), "19:02:00"},
		{"07:03", New(7, 3, 0), "7:03:00"},
		{"07:04:02", New(7, 4, 2), "7:04:02"},
		{"19:04:30", New(19, 4, 30), "19:04:30"},
	} {
		if res, err := Parse(testcase.s); err != nil {
			t.Error(err)
		} else if res != testcase.expected {
			t.Errorf("%s expected %v; got %v", testcase.s, testcase.expected, res)
		} else if res.String() != testcase.str {
			t.Errorf("#%d: got %q; want %q", i, res.String(), testcase.str)
		}
	}
	for _, testcase := range []string{
		"",
		"abc",
		"24:00",
	} {
		if _, err := Parse(testcase); err == nil {
			t.Errorf("%s expected error; got nil", testcase)
		}
	}
}

func TestSeconds(t *testing.T) {
	for i, testcase := range []struct {
		c        Clock
		expected int
	}{
		{New(0, 0, 0), 0},
		{New(7, 1, 2), 7*secondsPerHour + 1*secondsPerMinute + 2},
		{New(19, 4, 30), 19*secondsPerHour + 4*secondsPerMinute + 30},
		{New(25, 4, 30), secondsPerHour + 4*secondsPerMinute + 30},
	} {
		if res := testcase.c.Seconds(); res != testcase.expected {
			t.Errorf("#%d: expected %d; got %d", i, testcase.expected, res)
		}
	}
}

func TestSub(t *testing.T) {
	for i, tc := range []struct {
		c Clock
		u Clock
		d time.Duration
	}{
		{Clock{}, Clock{}, 0},
		{New(0, 0, 1), Clock{}, time.Second},
		{Clock{}, New(0, 0, 1), -time.Second},
		{New(6, 5, 4), Clock{}, 6*time.Hour + 5*time.Minute + 4*time.Second},
		{New(1, 0, 0), New(0, 30, 0), 30 * time.Minute},
		{New(12, 0, 0), New(12, 30, 0), -30 * time.Minute},
	} {
		if got := tc.c.Sub(tc.u); got != tc.d {
			t.Errorf("#%d: Sub(%v, %v): got %v; want %v", i, tc.c, tc.u, got, tc.d)
		}
	}
}

func TestUntil(t *testing.T) {
	for i, tc := range []struct {
		c Clock
		u Clock
		d time.Duration
	}{
		{Clock{}, Clock{}, 0},
		{Clock{}, New(0, 0, 1), time.Second},
		{New(0, 0, 1), Clock{}, 23*time.Hour + 59*time.Minute + 59*time.Second},
		{Clock{}, New(6, 5, 4), 6*time.Hour + 5*time.Minute + 4*time.Second},
		{New(0, 30, 0), New(1, 0, 0), 30 * time.Minute},
		{New(12, 30, 0), New(12, 0, 0), 23*time.Hour + 30*time.Minute},
	} {
		if got := tc.c.Until(tc.u); got != tc.d {
			t.Errorf("#%d: Sub(%v, %v): got %v; want %v", i, tc.c, tc.u, got, tc.d)
		}
	}
}
