package scheduler

import (
	"testing"
	"time"
)

func TestScheduleNext(t *testing.T) {
	format := "2006/01/02 15:04:05"
	parse := func(s string) time.Time {
		res, err := time.Parse(format, s)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	for i, testcase := range []struct {
		s    Schedule
		t    string
		next string
	}{
		{NewSchedule(2000, 1, 1, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:00"},
		{NewSchedule(2000, 1, 1, nil), "1900/12/15 12:00:00", "2000/01/01 00:00:00"},
		{NewSchedule(2000, 1, 1, nil), "2000/01/15 12:00:00", "0001/01/01 00:00:00"},

		{NewSchedule(0, 1, 1, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:00"},
		{NewSchedule(0, 1, 1, nil), "1900/12/15 12:00:00", "1901/01/01 00:00:00"},
		{NewSchedule(0, 1, 1, nil), "2000/01/15 12:00:00", "2001/01/01 00:00:00"},

		{NewSchedule(2000, 0, 1, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:00"},
		{NewSchedule(2000, 0, 1, nil), "1900/12/15 12:00:00", "2000/01/01 00:00:00"},
		{NewSchedule(2000, 0, 1, nil), "2000/01/15 12:00:00", "2000/02/01 00:00:00"},
		{NewSchedule(2000, 0, 1, nil), "2000/12/15 12:00:00", "0001/01/01 00:00:00"},

		{NewSchedule(2000, 1, 0, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:00"},
		{NewSchedule(2000, 1, 0, nil), "1900/12/15 12:00:00", "2000/01/01 00:00:00"},
		{NewSchedule(2000, 1, 0, nil), "2000/02/15 12:00:00", "0001/01/01 00:00:00"},

		{NewSchedule(0, 0, 1, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:00"},
		{NewSchedule(0, 0, 1, nil), "1900/12/15 12:00:00", "1901/01/01 00:00:00"},
		{NewSchedule(0, 0, 1, nil), "2000/01/15 12:00:00", "2000/02/01 00:00:00"},

		{NewSchedule(2000, 0, 0, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:00"},
		{NewSchedule(2000, 0, 0, nil), "1900/12/15 12:00:00", "2000/01/01 00:00:00"},
		{NewSchedule(2000, 0, 0, nil), "2001/01/15 12:00:00", "0001/01/01 00:00:00"},

		{NewSchedule(0, 1, 0, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:00"},
		{NewSchedule(0, 1, 0, nil), "1900/12/15 12:00:00", "1901/01/01 00:00:00"},
		{NewSchedule(0, 1, 0, nil), "2000/02/15 12:00:00", "2001/01/01 00:00:00"},

		{NewSchedule(0, 0, 0, nil), "2000/01/01 12:00:00", "2000/01/01 12:00:00"},
		{NewSchedule(0, 0, 0, nil), "1900/12/15 12:00:00", "1900/12/15 12:00:00"},
		{NewSchedule(0, 0, 0, nil), "2000/02/15 12:00:00", "2000/02/15 12:00:00"},
	} {
		if res := testcase.s.Next(parse(testcase.t)).Format(format); res != testcase.next {
			t.Errorf("#%d expected %v; got %v", i, testcase.next, res)
		} else if next := parse(testcase.next); !next.IsZero() && !testcase.s.IsMatched(next) {
			t.Errorf("#%d expected matched; got not", i)
		}
	}
}
