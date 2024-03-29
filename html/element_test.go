package html

import "testing"

func TestElement(t *testing.T) {
	for _, tc := range []struct {
		tag     string
		attrs   [][2]string
		content string
		html    HTML
	}{
		{"a", [][2]string{{"href", "/"}, {"style", "display:none"}}, "test", `<a href="/" style="display:none">test</a>`},
		{"p", nil, "test", "<p>test</p>"},
		{"p", [][2]string{{"hidden", "true"}}, "test", "<p hidden>test</p>"},
		{"p", [][2]string{{"hidden", "false"}}, "test", "<p>test</p>"},
		{"div", nil, "<test>", "<div>&lt;test&gt;</div>"},
	} {
		e := NewElement(tc.tag).Content(tc.content)
		for _, i := range tc.attrs {
			e.Attribute(i[0], i[1])
		}
		if res := e.HTML(); tc.html != res {
			t.Errorf("expected %q; got %q", tc.html, res)
		}
	}
	if div, expect := Div().Content(Br()).String(), "<div><br></div>"; expect != div {
		t.Errorf("expected %q; got %q", expect, div)
	}
}

func TestAppend(t *testing.T) {
	e := Div()
	if expect := "<div></div>"; expect != e.String() {
		t.Errorf("expected %q; got %q", expect, e.String())
	}
	e.AppendContent("test")
	if expect := "<div>test</div>"; expect != e.String() {
		t.Errorf("expected %q; got %q", expect, e.String())
	}
	e.AppendChild(Img().Src("test"))
	if expect := `<div>test<img src="test"></div>`; expect != e.String() {
		t.Errorf("expected %q; got %q", expect, e.String())
	}
}

func TestBackground(t *testing.T) {
	e := Background()
	if expect := ""; expect != e.String() {
		t.Errorf("expected %q; got %q", expect, e.String())
	}
	e.AppendContent("test")
	if expect := "test"; expect != e.String() {
		t.Errorf("expected %q; got %q", expect, e.String())
	}
	e.AppendChild(Img().Src("test"))
	if expect := `test<img src="test">`; expect != e.String() {
		t.Errorf("expected %q; got %q", expect, e.String())
	}
}
