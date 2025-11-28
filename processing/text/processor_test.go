package text

import (
	"regexp"
	"testing"
)

func TestRegexpRemover(t *testing.T) {
	for i, testcase := range []struct {
		re       *regexp.Regexp
		s        string
		expected string
	}{
		{regexp.MustCompile(""), "", ""},
		{regexp.MustCompile(`\d+`), "abc123", "abc"},
		{regexp.MustCompile(`\d+$`), "123abc456", "123abc"},
	} {
		if res, err := NewTasks().Append(RegexpRemover{testcase.re}).Process(testcase.s); err != nil {
			t.Error(err)
		} else if res != testcase.expected {
			t.Errorf("#%d: got %q; want %q", i, res, testcase.expected)
		}
	}
}

func TestCutter(t *testing.T) {
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
		if res, err := NewTasks().Append(Cutter{testcase.seq}).Process(testcase.s); err != nil {
			t.Error(err)
		} else if res != testcase.expected {
			t.Errorf("#%d: got %q; want %q", i, res, testcase.expected)
		}
	}
}

func TestTrimmer(t *testing.T) {
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
		if res, err := NewTasks().Append(Trimmer{testcase.cutset}).Process(testcase.s); err != nil {
			t.Error(err)
		} else if res != testcase.expected {
			t.Errorf("#%d: got %q; want %q", i, res, testcase.expected)
		}
	}
}

func TestLineToParagraph(t *testing.T) {
	for i, tc := range []struct {
		proc     Processor
		input    string
		expected string
	}{
		// Default behaviour (zero value): SkipEmpty=true, TrimSpace=false
		{
			LineToParagraph{},
			"  First line  \n\n  Second line\t \n\n",
			"<p>  First line  </p>\n<p></p>\n<p>  Second line\t </p>\n<p></p>\n",
		},
		// Explicitly enable trimming of leading/trailing whitespace
		{
			LineToParagraph{TrimSpace: true},
			"  Hello   \n  World  \n",
			"<p>Hello</p>\n<p>World</p>\n",
		},
		// Preserve empty lines (SkipEmpty = false)
		{
			LineToParagraph{SkipEmpty: false},
			"Line 1\n\n\nLine 2\n",
			"<p>Line 1</p>\n<p></p>\n<p></p>\n<p>Line 2</p>\n",
		},
		// Trim + preserve empty lines
		{
			LineToParagraph{TrimSpace: true, SkipEmpty: false},
			"   \n  A  \n   \nB   \n",
			"<p></p>\n<p>A</p>\n<p></p>\n<p>B</p>\n",
		},
		// Fully literal mode: keep all original whitespace and emit every line
		{
			LineToParagraph{TrimSpace: false, SkipEmpty: false},
			"\tIndented\n  \n    Spaces only   \n\nTrailing \n",
			"<p>\tIndented</p>\n<p>  </p>\n<p>    Spaces only   </p>\n<p></p>\n<p>Trailing </p>\n",
		},
		// Empty input
		{
			LineToParagraph{},
			"",
			"",
		},
		// Input containing only empty lines and whitespace
		{
			LineToParagraph{},
			"\n   \n\t\n  \n",
			"<p></p>\n<p>   </p>\n<p>\t</p>\n<p>  </p>\n",
		},
		// HTML escaping works regardless of configuration
		{
			LineToParagraph{TrimSpace: true},
			" <script>alert(1)</script> \n &copy; 2025 \n",
			"<p>&lt;script&gt;alert(1)&lt;/script&gt;</p>\n<p>&amp;copy; 2025</p>\n",
		},
		{
			LineToParagraph{TrimSpace: false},
			"  <b>bold</b>  \n",
			"<p>  &lt;b&gt;bold&lt;/b&gt;  </p>\n",
		},
	} {
		res, err := NewTasks().Append(tc.proc).Process(tc.input)
		if err != nil {
			t.Errorf("case %d: unexpected error: %v", i, err)
			continue
		}
		if res != tc.expected {
			t.Errorf("case %d:\ngot:\n%q\nwant:\n%q", i, res, tc.expected)
		}
	}
}
