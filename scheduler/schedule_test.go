package scheduler

import (
	"testing"
	"time"

	"github.com/sunshineplan/utils"
)

const format = "2006/01/02 15:04:05"

func parse(s string) time.Time {
	t, err := time.Parse(format, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestScheduleNext(t *testing.T) {
	for i, testcase := range []struct {
		s    Schedule
		t    string
		next string
	}{
		{NewSchedule(2000, 1, 1, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:01"},
		{NewSchedule(2000, 1, 1, nil), "1900/12/15 12:00:00", "2000/01/01 00:00:00"},
		{NewSchedule(2000, 1, 1, nil), "2000/01/15 12:00:00", "0001/01/01 00:00:00"},

		{NewSchedule(0, 1, 1, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:01"},
		{NewSchedule(0, 1, 1, nil), "1900/12/15 12:00:00", "1901/01/01 00:00:00"},
		{NewSchedule(0, 1, 1, nil), "2000/01/15 12:00:00", "2001/01/01 00:00:00"},

		{NewSchedule(2000, 0, 1, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:01"},
		{NewSchedule(2000, 0, 1, nil), "1900/12/15 12:00:00", "2000/01/01 00:00:00"},
		{NewSchedule(2000, 0, 1, nil), "2000/01/15 12:00:00", "2000/02/01 00:00:00"},
		{NewSchedule(2000, 0, 1, nil), "2000/12/15 12:00:00", "0001/01/01 00:00:00"},

		{NewSchedule(2000, 1, 0, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:01"},
		{NewSchedule(2000, 1, 0, nil), "1900/12/15 12:00:00", "2000/01/01 00:00:00"},
		{NewSchedule(2000, 1, 0, nil), "2000/02/15 12:00:00", "0001/01/01 00:00:00"},

		{NewSchedule(0, 0, 1, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:01"},
		{NewSchedule(0, 0, 1, nil), "1900/12/15 12:00:00", "1901/01/01 00:00:00"},
		{NewSchedule(0, 0, 1, nil), "2000/01/15 12:00:00", "2000/02/01 00:00:00"},

		{NewSchedule(2000, 0, 0, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:01"},
		{NewSchedule(2000, 0, 0, nil), "1900/12/15 12:00:00", "2000/01/01 00:00:00"},
		{NewSchedule(2000, 0, 0, nil), "2001/01/15 12:00:00", "0001/01/01 00:00:00"},

		{NewSchedule(0, 1, 0, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:01"},
		{NewSchedule(0, 1, 0, nil), "1900/12/15 12:00:00", "1901/01/01 00:00:00"},
		{NewSchedule(0, 1, 0, nil), "2000/02/15 12:00:00", "2001/01/01 00:00:00"},

		{NewSchedule(0, 0, 0, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:01"},
		{NewSchedule(0, 0, 0, nil), "1900/12/15 12:00:00", "1900/12/15 12:00:01"},
		{NewSchedule(0, 0, 0, nil), "2000/02/15 12:00:00", "2000/02/15 12:00:01"},
	} {
		if res := testcase.s.Next(parse(testcase.t)).Format(format); res != testcase.next {
			t.Errorf("#%d expected %v; got %v", i, testcase.next, res)
		} else if next := parse(testcase.next); !next.IsZero() && !testcase.s.IsMatched(next) {
			t.Errorf("#%d expected matched; got not %v", i, res)
		}
	}
}

func TestISOWeekScheduleNext(t *testing.T) {
	for i, testcase := range []struct {
		s    Schedule
		t    string
		next string
	}{
		{ISOWeekSchedule(2000, 1, utils.Ptr(time.Monday), nil), "2000/01/03 12:00:00", "2000/01/03 12:00:01"},
		{ISOWeekSchedule(2000, 1, utils.Ptr(time.Monday), nil), "1900/12/15 12:00:00", "2000/01/03 00:00:00"},
		{ISOWeekSchedule(2000, 1, utils.Ptr(time.Monday), nil), "2000/01/15 12:00:00", "0001/01/01 00:00:00"},

		{ISOWeekSchedule(0, 1, utils.Ptr(time.Monday), nil), "2000/01/03 12:00:00", "2000/01/03 12:00:01"},
		{ISOWeekSchedule(0, 1, utils.Ptr(time.Monday), nil), "1900/12/15 12:00:00", "1900/12/31 00:00:00"},
		{ISOWeekSchedule(0, 1, utils.Ptr(time.Monday), nil), "2000/01/15 12:00:00", "2001/01/01 00:00:00"},

		{ISOWeekSchedule(2000, 0, utils.Ptr(time.Monday), nil), "2000/01/03 12:00:00", "2000/01/03 12:00:01"},
		{ISOWeekSchedule(2000, 0, utils.Ptr(time.Monday), nil), "2000/01/01 12:00:00", "2000/01/03 00:00:00"},
		{ISOWeekSchedule(2000, 0, utils.Ptr(time.Monday), nil), "2000/01/05 12:00:00", "2000/01/10 00:00:00"},

		{ISOWeekSchedule(2000, 1, nil, nil), "2000/01/03 12:00:00", "2000/01/03 12:00:01"},
		{ISOWeekSchedule(2000, 1, nil, nil), "2000/01/02 12:00:00", "2000/01/03 00:00:00"},
		{ISOWeekSchedule(2000, 1, nil, nil), "2000/01/10 12:00:00", "0001/01/01 00:00:00"},

		{ISOWeekSchedule(0, 0, utils.Ptr(time.Monday), nil), "2000/01/03 12:00:00", "2000/01/03 12:00:01"},
		{ISOWeekSchedule(0, 0, utils.Ptr(time.Monday), nil), "1900/12/15 12:00:00", "1900/12/17 00:00:00"},
		{ISOWeekSchedule(0, 0, utils.Ptr(time.Monday), nil), "2000/01/15 12:00:00", "2000/01/17 00:00:00"},

		{ISOWeekSchedule(2000, 0, nil, nil), "2000/01/03 12:00:00", "2000/01/03 12:00:01"},
		{ISOWeekSchedule(2000, 0, nil, nil), "2000/01/01 12:00:00", "2000/01/03 00:00:00"},
		{ISOWeekSchedule(2000, 0, nil, nil), "2001/01/15 12:00:00", "0001/01/01 00:00:00"},

		{ISOWeekSchedule(0, 1, nil, nil), "2000/01/03 12:00:00", "2000/01/03 12:00:01"},
		{ISOWeekSchedule(0, 1, nil, nil), "2000/01/01 12:00:00", "2000/01/03 00:00:00"},
		{ISOWeekSchedule(0, 1, nil, nil), "2000/01/10 12:00:00", "2001/01/01 00:00:00"},

		{ISOWeekSchedule(0, 0, nil, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:01"},
		{ISOWeekSchedule(0, 0, nil, nil), "1900/12/15 12:00:00", "1900/12/15 12:00:01"},
		{ISOWeekSchedule(0, 0, nil, nil), "2000/02/15 12:00:00", "2000/02/15 12:00:01"},
	} {
		if res := testcase.s.Next(parse(testcase.t)).Format(format); res != testcase.next {
			t.Errorf("#%d expected %v; got %v", i, testcase.next, res)
		} else if next := parse(testcase.next); !next.IsZero() && !testcase.s.IsMatched(next) {
			t.Errorf("#%d expected matched; got not %v", i, res)
		}
	}
}

func TestTickerScheduleNext(t *testing.T) {
	for i, testcase := range []struct {
		d     time.Duration
		start string
		t     string
		next  string
	}{
		{time.Second, "2000/01/01 00:00:00", "2000/01/01 00:00:00", "2000/01/01 00:00:01"},
		{time.Minute, "2000/01/01 00:00:00", "2000/01/01 00:00:00", "2000/01/01 00:01:00"},
		{time.Minute, "2000/01/01 00:00:00", "2000/01/01 00:00:30", "2000/01/01 00:01:00"},
	} {
		s := &tickerSched{testcase.d, parse(testcase.start)}
		if res := s.Next(parse(testcase.t)).Format(format); res != testcase.next {
			t.Errorf("#%d expected %v; got %v", i, testcase.next, res)
		}
	}
}
