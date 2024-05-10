package text

import (
	"regexp"
	"testing"
)

func TestRemoveByRegexp(t *testing.T) {
	for i, testcase := range []struct {
		re       *regexp.Regexp
		s        string
		expected string
	}{
		{regexp.MustCompile(""), "", ""},
		{regexp.MustCompile(`\d+`), "abc123", "abc"},
		{regexp.MustCompile(`\d+$`), "123abc456", "123abc"},
	} {
		if res, err := NewTasks().Append(RemoveByRegexp{testcase.re}).Process(testcase.s); err != nil {
			t.Error(err)
		} else if res != testcase.expected {
			t.Errorf("#%d: got %q; want %q", i, res, testcase.expected)
		}
	}
}

func TestCut(t *testing.T) {
	for i, testcase := range []struct {
		seq      string
		s        string
		expected string
	}{
		{"", "", ""},
		{" ", "abc 123", "abc"},
		{" ", " abc 123", ""},
		{"abc", "123abc456", "123"},
	} {
		if res, err := NewTasks().Append(Cut{testcase.seq}).Process(testcase.s); err != nil {
			t.Error(err)
		} else if res != testcase.expected {
			t.Errorf("#%d: got %q; want %q", i, res, testcase.expected)
		}
	}
}

func TestTrim(t *testing.T) {
	for i, testcase := range []struct {
		cutset   string
		s        string
		expected string
	}{
		{"", "", ""},
		{" ", " abc 123 ", "abc 123"},
		{" ", " abc 123\n", "abc 123\n"},
		{" \n", " abc 123\n", "abc 123"},
	} {
		if res, err := NewTasks().Append(Trim{testcase.cutset}).Process(testcase.s); err != nil {
			t.Error(err)
		} else if res != testcase.expected {
			t.Errorf("#%d: got %q; want %q", i, res, testcase.expected)
		}
	}
}
