package text

import "testing"

func TestTrimSpace(t *testing.T) {
	task := NewTasks(TrimSpace())
	for i, testcase := range []struct {
		s        string
		expected string
	}{
		{"", ""},
		{" abc", "abc"},
		{"abc\n", "abc"},
		{"a b c", "a b c"},
	} {
		if res, err := task.Process(testcase.s); err != nil {
			t.Error(err)
		} else if res != testcase.expected {
			t.Errorf("#%d: got %q; want %q", i, res, testcase.expected)
		}
	}
}

func TestCutSpace(t *testing.T) {
	task := NewTasks(CutSpace())
	for i, testcase := range []struct {
		s        string
		expected string
	}{
		{"", ""},
		{" abc", "abc"},
		{"abc def", "abc"},
		{"abc\ndef", "abc"},
	} {
		if res, err := task.Process(testcase.s); err != nil {
			t.Error(err)
		} else if res != testcase.expected {
			t.Errorf("#%d: got %q; want %q", i, res, testcase.expected)
		}
	}
}

func TestRemoveParentheses(t *testing.T) {
	task := NewTasks(RemoveParentheses())
	for i, testcase := range []struct {
		s        string
		expected string
	}{
		{"", ""},
		{"abc", "abc"},
		{"abc(123)", "abc"},
		{"abc（123）", "abc"},
		{"abc（123)", "abc（123)"},
		{"abc(123)def", "abcdef"},
	} {
		if res, err := task.Process(testcase.s); err != nil {
			t.Error(err)
		} else if res != testcase.expected {
			t.Errorf("#%d: got %q; want %q", i, res, testcase.expected)
		}
	}
}

func TestTasks(t *testing.T) {
	task := NewTasks(TrimSpace(), RemoveParentheses(), CutSpace())
	for i, testcase := range []struct {
		s        string
		expected string
	}{
		{"", ""},
		{" abc(123)\n", "abc"},
		{"abc (123)", "abc"},
		{"(123)abc", "abc"},
	} {
		if res, err := task.Process(testcase.s); err != nil {
			t.Error(err)
		} else if res != testcase.expected {
			t.Errorf("#%d: got %q; want %q", i, res, testcase.expected)
		}
	}
}
