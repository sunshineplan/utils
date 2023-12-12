package html

import "testing"

func TestAttributes(t *testing.T) {
	defer func() {
		if e := recover(); e == nil {
			t.Error("expected panic")
		}
	}()
	Attributes("test")
}

func TestElement(t *testing.T) {
	for _, tc := range []struct {
		tag     string
		attrs   []string
		content string
		html    HTML
	}{
		{"a", []string{"href", "/", "style", "display:none"}, "test", `<a href="/" style="display:none">test</a>`},
		{"p", nil, "test", "<p>test</p>"},
		{"p", []string{"hidden", ""}, "test", "<p hidden>test</p>"},
		{"div", nil, "<test>", "<div>&lt;test&gt;</div>"},
	} {
		if res := Element(tc.tag, Attributes(tc.attrs...), tc.content); tc.html != res {
			t.Errorf("expected %q; got %q", tc.html, res)
		}
	}
	if div, expect := Element("div", nil, HTML("<br>")), "<div><br></div>"; expect != string(div) {
		t.Errorf("expected %q; got %q", expect, div)
	}
}
